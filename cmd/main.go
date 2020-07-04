package main

import (
	"context"
	"fmt"
	"io"
	"os"

	gcp_cloudsql_slowquery_dump "github.com/mtanda/gcp-cloudsql-slowquery-dump"
)

const (
	exitFail = 1
)

func run(args []string, stdout io.Writer) error {
	ctx := context.Background()
	return gcp_cloudsql_slowquery_dump.DumpSlowQuery(ctx, gcp_cloudsql_slowquery_dump.GCSEvent{Bucket: args[1], Name: args[2]})
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
