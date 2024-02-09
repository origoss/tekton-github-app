package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	tektonapi "github.com/origoss/tekton-github-app/pkg/tekton-api"
)

func main() {
	slog.Info("tekton github app client invoked")
	event := &tektonapi.TektonEvent{
		CheckSuite: tektonapi.CheckSuite{
			RepoOwner: os.Getenv("REPO_OWNER"),
			RepoName:  os.Getenv("REPO_NAME"),
			HeadSHA:   os.Getenv("HEAD_SHA"),
		},
		CheckRun: tektonapi.CheckRun{
			Name:    os.Getenv("NAME"),
			Title:   os.Getenv("TITLE"),
			Summary: os.Getenv("SUMMARY"),
		},
	}
	command := os.Getenv("COMMAND")
	switch strings.ToLower(command) {
	case "create_checkrun":
		event.Type = tektonapi.TektonEventCreateCheckRun
	case "update_checkrun":
		event.Type = tektonapi.TektonEventUpdateCheckRun
	default:
		slog.Error("invalid configuration", "COMMAND", command)
		panic(nil)
	}
	if conclusion := os.Getenv("CONCLUSION"); conclusion != "" {
		parsed, err := tektonapi.ParseCheckRunConclusion(conclusion)
		if err != nil {
			slog.Error("invalid configuration",
				"CONCLUSION", conclusion,
				"err", err)
			panic(err)
		}
		event.CheckRun.Conclusion = &parsed
	}
	status, err := tektonapi.ParseCheckRunStatus(os.Getenv("STATUS"))
	if err != nil {
		slog.Error("invalid configuration",
			"STATUS", status,
			"err", err)
		panic(err)
	}
	event.CheckRun.Status = status
	if id := os.Getenv("ID"); id != "" {
		idnum, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("invalid configuration",
				"ID", id,
				"err", err)
			panic(err)
		}
		event.CheckRun.ID = idnum
	}

	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	err = encoder.Encode(event)
	if err != nil {
		slog.Error("error encoding event to json",
			"event", event,
			"err", err,
		)
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, os.Getenv("GHAPP_URL"), buffer)
	if err != nil {
		slog.Error("error creating new request",
			"err", err,
		)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	slog.Debug("Sending event request to ghapp")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error sending event to ghapp",
			"err", err,
			"GHAPP_URL", os.Getenv("GHAPP_URL"),
		)
		panic(err)
	}
	if resp.StatusCode > 299 {
		slog.Error("invalid status code received from ghapp",
			"status", resp.Status,
			"body", resp.Body)
		panic(nil)
	}
}
