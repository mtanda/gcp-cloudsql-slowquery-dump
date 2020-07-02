package main

import (
	"io"
)

// GCS trigger event
type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
	Size   string `json:"size"`
}

// Cloud Logging data
type Entry struct {
	InsertID         string
	LogName          string
	ReceiveTimestamp string
	Resource         Resource
	TextPayload      string
	Timestamp        string
}

type Resource struct {
	Labels Labels
}

type Labels struct {
	ProjectId  string `json:"project_id"`
	Region     string
	DatabaseId string `json:"database_id"`
}

type SlowQuerySource struct {
	Labels
	Reader io.Reader
}
