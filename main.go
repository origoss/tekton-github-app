package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v58/github"
)

var logLevel = new(slog.LevelVar)

const (
	port_name                    = "PORT"
	gh_app_private_key_path_name = "GH_APP_PRIVATE_KEY_PATH"
	gh_app_id_name               = "GH_APP_ID"
	gh_app_installation_id_name  = "GH_APP_INSTALLATION_ID"
	gh_app_webhook_secret_name   = "GH_APP_WEBHOOK_SECRET"
)

type ghApp struct {
	privateKeyPath string
	appID          int64
	installationID int64
	webhookSecret  string
}

type config struct {
	httpServerPort int
	ghApp
}

func (c *config) logValues() {
	slog.Debug("configuration parsed", port_name, c.httpServerPort)
	slog.Debug("configuration parsed", gh_app_private_key_path_name, c.ghApp.privateKeyPath)
	slog.Debug("configuration parsed", gh_app_id_name, c.ghApp.appID)
	slog.Debug("configuration parsed", gh_app_installation_id_name, c.ghApp.installationID)
	slog.Debug("configuration parsed", gh_app_webhook_secret_name, c.ghApp.webhookSecret)
}

func init() {
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logLevel.Set(slog.LevelDebug)
	}
}

func getConfig() (*config, error) {
	slog.Debug("Parsing environment variables")
	httpServerPort, err := strconv.Atoi(os.Getenv(port_name))
	if err != nil {
		return nil, fmt.Errorf("Cannot parse the %s environment variable to int: %w", port_name, err)
	}
	appID, err := strconv.ParseInt(os.Getenv(gh_app_id_name), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse the %s environment variable to int: %w", gh_app_id_name, err)
	}
	installationID, err := strconv.ParseInt(os.Getenv(gh_app_installation_id_name), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse the %s environment variable to int: %w", gh_app_installation_id_name, err)
	}
	return &config{
		httpServerPort: httpServerPort,
		ghApp: ghApp{
			privateKeyPath: os.Getenv(gh_app_private_key_path_name),
			appID:          appID,
			installationID: installationID,
			webhookSecret:  os.Getenv(gh_app_webhook_secret_name),
		},
	}, nil
}

func checkPipelineCompleted(client *github.Client, repo *github.Repository, checkRun *github.CheckRun) {
	slog.Debug("checkPipelineCompleted")
	time.Sleep(5 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	status := "completed"
	conclusion := "success"
	title := "Tekton CI check completed"
	summary := "Tekton CI check in progress"
	_, _, err := client.Checks.UpdateCheckRun(ctx,
		*repo.Owner.Login,
		*repo.Name,
		checkRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   "Tekton CI check",
			Status: &status,
			Conclusion: &conclusion,
			Output: &github.CheckRunOutput{
				Title:   &title,
				Summary: &summary,
			},
		},
	)
	if err != nil {
		slog.Error("cannot update checkrun", "err", err)
	}
}



func checkPipelineInProgress(client *github.Client, repo *github.Repository, checkRun *github.CheckRun) {
	slog.Debug("checkPipelineInProgress")
	time.Sleep(5 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	status := "in_progress"
	title := "Tekton CI check"
	summary := "Tekton CI check in progress"
	checkRun, _, err := client.Checks.UpdateCheckRun(ctx,
		*repo.Owner.Login,
		*repo.Name,
		checkRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   "Tekton CI check",
			Status: &status,
			Output: &github.CheckRunOutput{
				Title:   &title,
				Summary: &summary,
			},
		},
	)
	if err != nil {
		slog.Error("cannot update checkrun", "err", err)
	}
	checkPipelineCompleted(client, repo, checkRun)
}

func checkPipeline(ctx context.Context, client *github.Client, repo *github.Repository, cSuite *github.CheckSuite) error {
	status := "queued"
	title := "Tekton CI check"
	summary := "Tekton CI summary"
	checkRun, _, err := client.Checks.CreateCheckRun(ctx,
		*repo.Owner.Login,
		*repo.Name,
		github.CreateCheckRunOptions{
			Name:      "Tekton CI check",
			HeadSHA:   *cSuite.HeadSHA,
			Status:    &status,
			StartedAt: &github.Timestamp{Time: time.Now()},
			Output: &github.CheckRunOutput{
				Title:   &title,
				Summary: &summary,
			},
		})
	if err != nil {
		return fmt.Errorf("cannot create checkrun: %w", err)
	}
	go checkPipelineInProgress(client, repo, checkRun)
	return nil
}

func main() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
	slog.Info("Starting tekton-github-app")
	conf, err := getConfig()
	if err != nil {
		slog.Error("Invalid configuration", "err", err)
		panic(err)
	}
	conf.logValues()

	tr := http.DefaultTransport

	itr, err := ghinstallation.NewKeyFromFile(tr,
		conf.ghApp.appID,
		conf.ghApp.installationID,
		conf.ghApp.privateKeyPath,
	)
	if err != nil {
		slog.Error("Cannot create GitHub transport", "err", err)
	}
	client := github.NewClient(&http.Client{Transport: itr})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("HTTP request received")
		payload, err := github.ValidatePayload(r, []byte(conf.ghApp.webhookSecret))
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
			err := checkPipeline(ctx, client, event.GetRepo(), event.CheckSuite)
			if err != nil {
				slog.Error("checkPipeline failed", "err", err)
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
	})
	err = http.ListenAndServe(fmt.Sprintf(":%d", conf.httpServerPort), nil)
	slog.Error("Cannot start HTTP server", "err", err)
}
