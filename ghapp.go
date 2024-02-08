package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v58/github"
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := checkPipeline(ctx, gh.client, event.GetRepo(), event.CheckSuite)
		if err != nil {
			slog.Error("checkPipeline failed", "err", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
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
