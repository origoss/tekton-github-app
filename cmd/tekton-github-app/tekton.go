package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/origoss/tekton-github-app/pkg/tekton-api"
)

// This file contains the code for maintaining the connection towards
// Tekton CI

type tektonConfig struct {
	tektonUrl string
	tektonUser string
	tektonPassword string
}

type tekton struct {
	conf *tektonConfig
	gh *gh
}

func (t *tekton) registerGH(gh *gh) {
	t.gh = gh
}

// This event comes from gh
func (t *tekton) handleCheckSuiteEvent(ctx context.Context, cs tektonapi.CheckSuite) error {
	slog.Debug("handleCheckSuiteEvent", "cs", cs)
	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(tektonapi.CheckSuiteCreatedBody{
		Event: "check-suite-created",
		CheckSuite: cs,
	})
	if err != nil {
		return fmt.Errorf("error encoding body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.conf.tektonUrl, buffer)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.conf.tektonUser != "" {
		req.SetBasicAuth(t.conf.tektonUser, t.conf.tektonPassword)
	}
	slog.Debug("sending request to tekton")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		slog.Error("invalid status code received from tekton",
			"status", resp.Status,
			"body", resp.Body)
		return fmt.Errorf("invalid status code from tekton")
	}
	return nil
}

func (t *tekton) handleTektonEvents(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handleTektonEvents invoked")
	decoder := json.NewDecoder(r.Body)
	event := tektonapi.TektonEvent{}
	err := decoder.Decode(&event)
	if err != nil {
		slog.Error("error decoding tekton event body", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.Debug("body parsed", "type", event.Type)
	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	switch event.Type {
	case tektonapi.TektonEventCreateCheckRun:
		id, err := t.gh.createCheckRun(ctx, event.CheckSuite, event.CheckRun)
		if err != nil {
			slog.Error("error creating checkrun", "err", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusCreated)
		bodyEncoder := json.NewEncoder(w)
		err = bodyEncoder.Encode(tektonapi.CheckRunCreatedResponseBody{
			ID: id,
		})
		if err != nil {
			slog.Error("error encoding response body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case tektonapi.TektonEventUpdateCheckRun:
		err := t.gh.updateCheckRun(ctx, event.CheckSuite, event.CheckRun)
		if err != nil {
			slog.Error("error updating checkrun", "err", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
}

func newTekton(conf *tektonConfig) *tekton {
	t := &tekton{
		conf: conf,
	}
	http.HandleFunc("/tekton/", t.handleTektonEvents)
	return t
}
