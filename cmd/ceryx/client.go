package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultAPIURL = "http://localhost:8080"

type apiError struct {
	Status int
	Detail string
}

func (e *apiError) Error() string { return fmt.Sprintf("%s (HTTP %d)", e.Detail, e.Status) }

type client struct {
	base string
	http *http.Client
}

func newClient(base string) *client {
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	return &client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 30 * time.Second}}
}

// do calls the API and returns the raw response body. Non-2xx responses become *apiError.
func (c *client) do(method, path string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		var e struct{ Detail string }
		if json.Unmarshal(data, &e) != nil || e.Detail == "" {
			e.Detail = strings.TrimSpace(string(data))
			if e.Detail == "" {
				e.Detail = http.StatusText(resp.StatusCode)
			}
		}
		return nil, &apiError{Status: resp.StatusCode, Detail: e.Detail}
	}
	return data, nil
}

func resource(kind, name string) string { return "/api/" + kind + "/" + url.PathEscape(name) }
