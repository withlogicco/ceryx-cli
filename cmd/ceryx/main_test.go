package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun(t *testing.T) {
	var method, path, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		method, path, body = r.Method, r.URL.EscapedPath(), string(b)
		if r.Method == "GET" {
			w.WriteHeader(404)
			w.Write([]byte(`{"detail":"Route not found"}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tests := []struct {
		args               []string
		code               int
		method, path, body string
	}{
		{[]string{"routes", "set", "a.com", "app:80", "--enforce-https"}, 0, "POST", "/api/routes",
			`{"settings":{"enforce_https":true},"source":"a.com","target":"app:80"}`},
		{[]string{"routes", "patch", "$wildcard", "--certificate-path", "", "--mtls-verify-depth", "2"}, 0, "PUT", "/api/routes/$wildcard",
			`{"settings":{"certificate_path":null,"mtls_verify_depth":2}}`},
		{[]string{"certs", "patch", "*.a.com", "--meta", `{"vendor":"x"}`}, 0, "PUT", "/api/certificates/%2A.a.com",
			`{"meta":{"vendor":"x"}}`},
		{[]string{"routes", "get", "missing.com"}, exitNotFound, "GET", "/api/routes/missing.com", ""},
		{[]string{"routes", "patch", "a.com"}, exitUsage, "", "", ""},
		{[]string{"routes", "update", "a.com"}, exitUsage, "", "", ""},
	}
	for _, tt := range tests {
		method, path, body = "", "", ""
		var stderr bytes.Buffer
		code := run(append(tt.args, "--api-url", srv.URL), io.Discard, &stderr)
		if code != tt.code || method != tt.method || path != tt.path || body != tt.body {
			t.Errorf("%v: got %d %s %s %s (%s)", tt.args, code, method, path, body, stderr.String())
		}
	}
}
