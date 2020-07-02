package main

import (
	"io"
	"text/template"
	"time"
)

type nopSeeker struct {
	io.Reader
}

func (nopSeeker) Seek(offset int64, whence int) (int64, error) { return 0, nil }

func NopSeeker(r io.Reader) io.ReadSeeker {
	return nopSeeker{r}
}

func fmtUnixTime(a time.Time) int64 {
	return a.Unix()
}

func fmtRFC3339(a time.Time) string {
	return a.Format("2006-01-02T15:04:05.999999Z")
}

func getTemplate() (*template.Template, error) {
	tpl := `# Time: {{fmtRFC3339 .Ts}}
# User@Host: {{.User}} @ {{.Host}}
# Query_time: {{printf "%f" .TimeMetrics.Query_time}}  Lock_time: {{printf "%f" .TimeMetrics.Lock_time}} Rows_sent: {{.NumberMetrics.Rows_sent}}  Rows_examined: {{.NumberMetrics.Rows_examined}}
SET timestamp={{fmtUnixTime .Ts}};
{{.Query}};
`
	funcs := template.FuncMap{
		"fmtRFC3339":  fmtRFC3339,
		"fmtUnixTime": fmtUnixTime,
	}

	t, err := template.New("").Funcs(funcs).Parse(tpl)
	if err != nil {
		return nil, err
	}
	return t, nil
}
