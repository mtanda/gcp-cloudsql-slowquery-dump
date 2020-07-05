# gcp-cloudsql-slowquery-dump
Dump MySQL slow query log from raw Cloud Logging data.

## Setup
Create Log sink for CloudSQL slow query log.
https://cloud.google.com/logging/docs/export/configure_export_v2

Set trigger for raw slow query log in GCS.
```
gsutil notification create -p 'cloudsql.googleapis.com/mysql-slow.log/' -f json gs://slowquery-log-location
gcloud functions deploy cloudsql-slowquery-dump --entry-point DumpSlowQuery --runtime go113 --trigger-event google.pubsub.topic.publish --trigger-resource pubsub-topic
```
