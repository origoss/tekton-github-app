package tektonapi

import (
	"encoding/json"
	"fmt"
)

type TektonEventType int

const (
	TektonEventInvalid TektonEventType = iota
	TektonEventCreateCheckRun
	TektonEventUpdateCheckRun
)

func ParseTektonEventType(s string) (TektonEventType, error) {
	switch s {
	case "create-checkrun":
		return TektonEventCreateCheckRun, nil
	case "update-checkrun":
		return TektonEventUpdateCheckRun, nil
	default:
		return TektonEventInvalid, fmt.Errorf("invalid TektonEventType: %s", s)
	}
}

func (tet TektonEventType) String() string {
	switch tet {
	case TektonEventCreateCheckRun:
		return "create-checkrun"
	case TektonEventUpdateCheckRun:
		return "update-checkrun"
	default:
		panic(fmt.Sprintf("invalid TektonEventType: %d", tet))
	}
}

func (tet TektonEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(tet.String())
}

func (tet *TektonEventType) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	if *tet, err = ParseTektonEventType(s); err != nil {
		return err
	}
	return nil
}

type CheckSuite struct {
	RepoOwner string `json:"repo-owner"`
	RepoName  string `json:"repo-name"`
	HeadSHA   string `json:"head-sha"`
}

type CheckRunStatus int

const (
	CheckRunStatusInvalid CheckRunStatus = iota
	CheckRunStatusQueued
	CheckRunStatusInProgress
	CheckRunStatusSuccess
)

func ParseCheckRunStatus(s string) (CheckRunStatus, error) {
	switch s {
	case "queued":
		return CheckRunStatusQueued, nil
	case "in_progress":
		return CheckRunStatusInProgress, nil
	case "success":
		return CheckRunStatusSuccess, nil
	default:
		return CheckRunStatusInvalid, fmt.Errorf("invalid CheckRunStatus: %s", s)
	}
}

func (crs CheckRunStatus) String() string {
	switch crs {
	case CheckRunStatusQueued:
		return "queued"
	case CheckRunStatusInProgress:
		return "in_progress"
	case CheckRunStatusSuccess:
		return "success"
	default:
		panic(fmt.Sprintf("invalid CheckRunStatus: %d", crs))
	}
}

func (crs CheckRunStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(crs.String())
}

func (crs *CheckRunStatus) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	if *crs, err = ParseCheckRunStatus(s); err != nil {
		return err
	}
	return nil
}

type CheckRunConclusion int

const (
	CheckRunConclusionInvalid CheckRunConclusion = iota
	CheckRunConclusionSuccess
	CheckRunConclusionFailure
	CheckRunConclusionNeutral
	CheckRunConclusionCancelled
	CheckRunConclusionSkipped
	CheckRunConclusionTimedOut
)

func ParseCheckRunConclusion(s string) (CheckRunConclusion, error) {
	switch s {
	case "success":
		return CheckRunConclusionSuccess, nil
	case "failure":
		return CheckRunConclusionFailure, nil
	case "neutral":
		return CheckRunConclusionNeutral, nil
	case "cancelled":
		return CheckRunConclusionCancelled, nil
	case "skipped":
		return CheckRunConclusionSkipped, nil
	case "timed_out":
		return CheckRunConclusionTimedOut, nil
	default:
		return CheckRunConclusionInvalid, fmt.Errorf("invalid CheckRunConclusion: %s", s)
	}
}

func (crc CheckRunConclusion) String() string {
	switch crc {
	case CheckRunConclusionSuccess:
		return "success"
	case CheckRunConclusionFailure:
		return "failure"
	case CheckRunConclusionNeutral:
		return "neutral"
	case CheckRunConclusionCancelled:
		return "cancelled"
	case CheckRunConclusionSkipped:
		return "skipped"
	case CheckRunConclusionTimedOut:
		return "timed_out"
	default:
		panic(fmt.Sprintf("invalid CheckRunConclusion: %d", crc))
	}
}

func (crc CheckRunConclusion) MarshalJSON() ([]byte, error) {
	return json.Marshal(crc.String())
}

func (crc *CheckRunConclusion) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	if *crc, err = ParseCheckRunConclusion(s); err != nil {
		return err
	}
	return nil
}

type CheckRun struct {
	Name       string              `json:"name"`
	Title      string              `json:"title"`
	Summary    string              `json:"summary"`
	Conclusion *CheckRunConclusion `json:"conclusion,omitempty"`
	Status     CheckRunStatus      `json:"status"`
	ID         int64               `json:"id"`
}

type TektonEvent struct {
	Type       TektonEventType `json:"type"`
	CheckSuite `json:"check-suite"`
	CheckRun   `json:"check-run"`
}

type CheckRunCreatedResponseBody struct {
	ID int64 `json:"id"`
}

type CheckSuiteCreatedBody struct {
	Event      string `json:"event"`
	CheckSuite `json:"check-suite"`
}
