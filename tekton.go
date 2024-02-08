package main

import (
	"context"
	"log/slog"
	"time"
)

// This file contains the code for maintaining the connection towards
// Tekton CI

type tekton struct {
	gh *gh
}

func (t *tekton) registerGH(gh *gh) {
	t.gh = gh
}

type checkSuite struct {
	cSuiteHeadSHA string
	repoOwner     string
	repoName      string
}

// This event comes from gh
func (t *tekton) handleCheckSuiteEvent(ctx context.Context, cs *checkSuite) error {
	// TODO: implement properly
	go func() {
		if t.gh != nil {
			time.Sleep(3 * time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			opts := &checkrunOpts{
				name:       "Tekton test CI",
				title:      "Queuing task",
				summary:    "Queuing test CI task",
				checkSuite: cs,
			}
			_, err := t.gh.createCheckRun(ctx, opts)
			if err != nil {
				slog.Error("cannot create checkrun", "err", err)
			}
			time.Sleep(5 * time.Second)
			opts.status = "in_progress"
			opts.title = "Processing task"
			opts.summary = "Working on it"
			err = t.gh.updateCheckRun(ctx, opts)
			if err != nil {
				slog.Error("cannot update checkrun", "err", err)
			}
			conclusion := "success"
			time.Sleep(5 * time.Second)
			opts.status = "completed"
			opts.conclusion = &conclusion
			opts.summary = "Task done"
			err = t.gh.updateCheckRun(ctx, opts)
			if err != nil {
				slog.Error("cannot update checkrun", "err", err)
			}
		}
	}()
	return nil
}

func newTekton() *tekton {
	t := &tekton{}
	return t
}
