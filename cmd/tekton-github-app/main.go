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
	tekton_user_name             = "TEKTON_USERNAME"
	tekton_password              = "TEKTON_PASSWORD"
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
	slog.Debug("configuration parsed", tekton_user_name, c.tektonConfig.tektonUser)
	slog.Debug("configuration parsed", tekton_password, c.tektonConfig.tektonPassword)
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
			tektonUser: os.Getenv(tekton_user_name),
			tektonPassword: os.Getenv(tekton_password),
		},
	}, nil
}

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
