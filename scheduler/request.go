package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

func validMethod(method string) bool {
	switch method {
	case "GET",
		"HEAD",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"CONNECT",
		"OPTIONS",
		"TRACE":
		return true
	}
	return false
}

type Replacement map[string]string

func (r Replacement) WithEnv() Replacement {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "ISUP_") {
			pair := strings.SplitN(e, "=", 2)
			key := strings.TrimPrefix(pair[0], "ISUP_")
			r[key] = pair[1]
		}
	}
	return r
}

type Request struct {
	// Insecure    bool
	Body        *string
	Headers     map[string]string
	Method      *string
	QueryParams map[string]string
	URL         *string
}

func (r *Request) Check() error {
	if r.URL == nil {
		return fmt.Errorf("URL cannot be empty")
	}
	if r.Method == nil {
		method := "GET"
		r.Method = &method
	} else {
		method := strings.ToUpper(*r.Method)
		r.Method = &method
	}
	if !validMethod(*r.Method) {
		return fmt.Errorf("Method '%s' isn't valid", *r.Method)
	}
	return nil
}

func (r *Request) Run(ctx context.Context, repl *Replacement) (*http.Response, error) {
	var body bytes.Buffer

	if r.Body != nil {
		if repl != nil {
			tmpl, err := template.New("body").Parse(*r.Body)
			if err != nil {
				return nil, err
			}
			err = tmpl.Execute(&body, repl)
			if err != nil {
				return nil, err
			}
		} else {
			body = *bytes.NewBuffer([]byte(*r.Body))
		}
	}

	req, err := http.NewRequest(*r.Method, *r.URL, &body)
	if err != nil {
		return nil, err
	}

	if len(r.QueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range r.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range r.Headers {
		req.Header.Add(k, v)
	}

	client := ctx.Value("http.client").(http.Client)
	req = req.WithContext(ctx)
	return client.Do(req)
}
