package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v58/github"
	"github.com/origoss/tekton-github-app/pkg/tekton-api"
)

// This file contains the code for maintaining the connection towards
// the GitHub API

type ghApp struct {
	privateKeyPath string
	appID          int64
	installationID int64
	webhookSecret  string
	httpServerPort int
}

type gh struct {
	client *github.Client
	conf   *ghApp
	tekton *tekton
}

func (gh *gh) registerTekton(t *tekton) {
	gh.tekton = t
}

func (gh *gh) createCheckRun(ctx context.Context, cs tektonapi.CheckSuite, cr tektonapi.CheckRun) (int64, error) {
	status := cr.Status.String()
	checkRun, _, err := gh.client.Checks.CreateCheckRun(ctx,
		cs.RepoOwner,
		cs.RepoName,
		github.CreateCheckRunOptions{
			HeadSHA:   cs.HeadSHA,
			Name:      cr.Name,
			Status:    &status,
			StartedAt: &github.Timestamp{Time: time.Now()},
			Output: &github.CheckRunOutput{
				Title:   &cr.Title,
				Summary: &cr.Summary,
			},
		})
	if err != nil {
		return -1, fmt.Errorf("cannot create checkrun: %w", err)
	}
	return checkRun.GetID(), nil
}

func (gh *gh) updateCheckRun(ctx context.Context, cs tektonapi.CheckSuite, cr tektonapi.CheckRun) error {
	status := cr.Status.String()
	var conclusion *string
	if cr.Conclusion != nil {
		c := cr.Conclusion.String()
		conclusion = &c
	}
	_, _, err := gh.client.Checks.UpdateCheckRun(ctx,
		cs.RepoOwner,
		cs.RepoName,
		cr.ID,
		github.UpdateCheckRunOptions{
			Name:       cr.Name,
			Status:     &status,
			Conclusion: conclusion,
			Output: &github.CheckRunOutput{
				Title:   &cr.Title,
				Summary: &cr.Summary,
			},
		})
	if err != nil {
		return fmt.Errorf("cannot update checkrun: %w", err)
	}
	return nil
}

func (gh *gh) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("HTTP request received")
	payload, err := github.ValidatePayload(r, []byte(gh.conf.webhookSecret))
	if err != nil {
		slog.Error("Invalid payload", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		slog.Error("Webhook event cannot be parsed", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Debug("event received", "event", event)
	switch event := event.(type) {
	case *github.CheckSuiteEvent:
		if gh.tekton != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			repo := event.GetRepo()
			err := gh.tekton.handleCheckSuiteEvent(ctx, tektonapi.CheckSuite{
				HeadSHA:   *event.GetCheckSuite().HeadSHA,
				RepoOwner: *repo.Owner.Login,
				RepoName:  *repo.Name,
			})
			if err != nil {
				slog.Error("error handling CheckSuite event", "err", err)
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
	}
}

func newGh(conf *ghApp) *gh {
	gh := &gh{
		conf: conf,
	}
	tr := http.DefaultTransport

	itr, err := ghinstallation.NewKeyFromFile(tr,
		conf.appID,
		conf.installationID,
		conf.privateKeyPath,
	)
	if err != nil {
		slog.Error("Cannot create GitHub transport", "err", err)
	}
	gh.client = github.NewClient(&http.Client{Transport: itr})

	http.Handle("/", gh)
	return gh
}
