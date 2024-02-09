package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var logLevel = new(slog.LevelVar)

const (
	port_name                    = "PORT"
	gh_app_private_key_path_name = "GH_APP_PRIVATE_KEY_PATH"
	gh_app_id_name               = "GH_APP_ID"
	gh_app_installation_id_name  = "GH_APP_INSTALLATION_ID"
	gh_app_webhook_secret_name   = "GH_APP_WEBHOOK_SECRET"
	tekton_url_name              = "TEKTON_URL"
)

type config struct {
	ghApp
	tektonConfig
}

func (c *config) logValues() {
	slog.Debug("configuration parsed", port_name, c.httpServerPort)
	slog.Debug("configuration parsed", gh_app_private_key_path_name, c.ghApp.privateKeyPath)
	slog.Debug("configuration parsed", gh_app_id_name, c.ghApp.appID)
	slog.Debug("configuration parsed", gh_app_installation_id_name, c.ghApp.installationID)
	slog.Debug("configuration parsed", gh_app_webhook_secret_name, c.ghApp.webhookSecret)
	slog.Debug("configuration parsed", tekton_url_name, c.tektonConfig.tektonUrl)
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
		ghApp: ghApp{
			privateKeyPath: os.Getenv(gh_app_private_key_path_name),
			appID:          appID,
			installationID: installationID,
			webhookSecret:  os.Getenv(gh_app_webhook_secret_name),
			httpServerPort: httpServerPort,
		},
		tektonConfig: tektonConfig{
			tektonUrl: os.Getenv(tekton_url_name),
		},
	}, nil
}

// func checkPipelineCompleted(client *github.Client, repo *github.Repository, checkRun *github.CheckRun) {
// 	slog.Debug("checkPipelineCompleted")
// 	time.Sleep(5 * time.Second)
// 	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 	defer cancel()
// 	status := "completed"
// 	conclusion := "success"
// 	title := "Tekton CI check completed"
// 	summary := "Tekton CI check in progress"
// 	_, _, err := client.Checks.UpdateCheckRun(ctx,
// 		*repo.Owner.Login,
// 		*repo.Name,
// 		checkRun.GetID(),
// 		github.UpdateCheckRunOptions{
// 			Name:       "Tekton CI check",
// 			Status:     &status,
// 			Conclusion: &conclusion,
// 			Output: &github.CheckRunOutput{
// 				Title:   &title,
// 				Summary: &summary,
// 			},
// 		},
// 	)
// 	if err != nil {
// 		slog.Error("cannot update checkrun", "err", err)
// 	}
// }

// func checkPipelineInProgress(client *github.Client, repo *github.Repository, checkRun *github.CheckRun) {
// 	slog.Debug("checkPipelineInProgress")
// 	time.Sleep(5 * time.Second)
// 	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 	defer cancel()
// 	status := "in_progress"
// 	title := "Tekton CI check"
// 	summary := "Tekton CI check in progress"
// 	checkRun, _, err := client.Checks.UpdateCheckRun(ctx,
// 		*repo.Owner.Login,
// 		*repo.Name,
// 		checkRun.GetID(),
// 		github.UpdateCheckRunOptions{
// 			Name:   "Tekton CI check",
// 			Status: &status,
// 			Output: &github.CheckRunOutput{
// 				Title:   &title,
// 				Summary: &summary,
// 			},
// 		},
// 	)
// 	if err != nil {
// 		slog.Error("cannot update checkrun", "err", err)
// 	}
// 	checkPipelineCompleted(client, repo, checkRun)
// }

// func checkPipeline(ctx context.Context, client *github.Client, repo *github.Repository, cSuite *github.CheckSuite) error {
// 	status := "queued"
// 	title := "Tekton CI check"
// 	summary := "Tekton CI summary"
// 	checkRun, _, err := client.Checks.CreateCheckRun(ctx,
// 		*repo.Owner.Login,
// 		*repo.Name,
// 		github.CreateCheckRunOptions{
// 			Name:      "Tekton CI check",
// 			HeadSHA:   *cSuite.HeadSHA,
// 			Status:    &status,
// 			StartedAt: &github.Timestamp{Time: time.Now()},
// 			Output: &github.CheckRunOutput{
// 				Title:   &title,
// 				Summary: &summary,
// 			},
// 		})
// 	if err != nil {
// 		return fmt.Errorf("cannot create checkrun: %w", err)
// 	}
// 	go checkPipelineInProgress(client, repo, checkRun)
// 	return nil
// }

func connect(gh *gh, t *tekton) {
	gh.registerTekton(t)
	t.registerGH(gh)
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

	gh := newGh(&conf.ghApp)
	t := newTekton(&conf.tektonConfig)
	connect(gh, t)
	err = http.ListenAndServe(fmt.Sprintf(":%d", conf.httpServerPort), nil)
	slog.Error("Cannot start HTTP server", "err", err)
}
