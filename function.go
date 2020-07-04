package gcp_cloudsql_slowquery_dump

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/percona/go-mysql/log"
	parser "github.com/percona/go-mysql/log/slow"
	"golang.org/x/sync/errgroup"
)

func newRawSlowQueryReader(ctx context.Context, client *storage.Client, bucket string, name string) (*storage.Reader, error) {
	reader, err := client.Bucket(bucket).Object(name).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func extractTextPayloadChan(src *storage.Reader, ch chan SlowQuerySource) {
	pipes := make(map[string]*io.PipeWriter)
	go func() {
		decoder := json.NewDecoder(src)
		for decoder.More() {
			var entry Entry
			if err := decoder.Decode(&entry); err != nil {
				continue // ignore error
			}

			uid := entry.Resource.Labels.ProjectId + "/" + entry.Resource.Labels.Region + "/" + entry.Resource.Labels.DatabaseId
			if _, ok := pipes[uid]; !ok {
				in, out := io.Pipe()
				pipes[uid] = out
				ch <- SlowQuerySource{
					Labels: entry.Resource.Labels,
					Reader: in,
				}
			}
			fmt.Fprintf(pipes[uid], "%s\n", entry.TextPayload)
		}
		for _, pipe := range pipes {
			pipe.Close()
		}
		close(ch)
	}()
}

func DumpSlowQuery(ctx context.Context, m PubSubMessage) error {
	var e GCSEvent
	err := json.Unmarshal(m.Data, &e)
	if err != nil {
		return err
	}

	srcBucket := e.Bucket
	srcObject := e.Name
	dstBucket := os.Getenv("DST_BUCKET")
	dstObjectPrefix := os.Getenv("DST_OBJECT_PREFIX")
	if dstBucket == "" {
		dstBucket = srcBucket
	}
	if len(dstObjectPrefix) == 0 {
		dstObjectPrefix = "mysql-slow.log/"
	} else if dstObjectPrefix[len(dstObjectPrefix)-1:len(dstObjectPrefix)] != "/" {
		dstObjectPrefix += "/"
	}
	excludeUsers := strings.Split(os.Getenv("EXCLUDE_USERS"), ",")
	excludeMap := make(map[string]bool)
	for _, u := range excludeUsers {
		excludeMap[u] = true
	}

	tpl, err := getTemplate()
	if err != nil {
		return err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	rsqReader, err := newRawSlowQueryReader(ctx, client, srcBucket, srcObject)
	if err != nil {
		return err
	}
	defer rsqReader.Close()

	ch := make(chan SlowQuerySource)
	extractTextPayloadChan(rsqReader, ch)
	eg := errgroup.Group{}
	for src := range ch {
		src := src
		eg.Go(func() error {
			reader := NopSeeker(src.Reader)
			p := parser.NewSlowLogParser(reader, log.Options{})
			if err != nil {
				return err
			}
			go p.Start()

			databaseId := src.DatabaseId[:len(src.ProjectId)]
			dstObject := dstObjectPrefix + src.ProjectId + "/" + src.Region + "/" + databaseId + "/" + strings.Replace(srcObject, ".json", ".log", 1)
			fmt.Fprintf(os.Stderr, "output to gs://%s/%s\n", dstBucket, dstObject)
			writer := client.Bucket(dstBucket).Object(dstObject).NewWriter(ctx)
			defer writer.Close()

			for e := range p.EventChan() {
				if _, exist := excludeMap[e.User]; exist {
					continue
				}
				tpl.Execute(writer, e)
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
