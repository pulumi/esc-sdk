// Copyright 2026, Pulumi Corporation.  All rights reserved.

// Package httpstub runs a local HTTP server that answers with a scripted
// response and records what it received, so tests can drive real clients end
// to end without a live Pulumi Cloud.
package httpstub

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   string
}

type Response struct {
	Status      int
	ContentType string
	Body        string
}

func JSON(status int, body string) Response {
	return Response{Status: status, ContentType: "application/json", Body: body}
}

func YAML(body string) Response {
	return Response{Status: http.StatusOK, ContentType: "application/x-yaml", Body: body}
}

func Empty(status int) Response {
	return Response{Status: status}
}

type Recorder struct {
	mu       sync.Mutex
	requests []Request
}

func (r *Recorder) record(req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, Request{
		Method: req.Method,
		Path:   req.URL.EscapedPath(),
		Query:  req.URL.Query(),
		Header: req.Header.Clone(),
		Body:   string(body),
	})
}

// Only returns the single request the server received and fails when the
// client sent none or more than one.
func (r *Recorder) Only(t *testing.T) Request {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	require.Len(t, r.requests, 1)
	return r.requests[0]
}

func (r *Recorder) All() []Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Request(nil), r.requests...)
}

// Serve starts a server that answers every request with response and returns
// its base URL plus the recorder of what arrived.
func Serve(t *testing.T, response Response) (string, *Recorder) {
	t.Helper()
	rec := &Recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		if response.ContentType != "" {
			w.Header().Set("Content-Type", response.ContentType)
		}
		w.WriteHeader(response.Status)
		_, _ = io.WriteString(w, response.Body)
	}))
	t.Cleanup(server.Close)
	return server.URL, rec
}

// Fixture returns a recorded Pulumi Cloud response from this package's
// testdata directory.
func Fixture(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", name))
	require.NoError(t, err)
	return string(content)
}

func RequireRequest(t *testing.T, got Request, method, path string, query url.Values) {
	t.Helper()
	require.Equal(t, method, got.Method)
	require.Equal(t, path, got.Path)
	if query == nil {
		query = url.Values{}
	}
	require.Equal(t, query, got.Query)
}

// DefaultExecutor sends a request with the default HTTP client, which is all a
// generated CloudClient needs to reach a stub server.
func DefaultExecutor(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// ServeRoutes starts a server that answers each request with the response
// registered for its "METHOD /path" and fails the test on any other request.
func ServeRoutes(t *testing.T, routes map[string]Response) (string, *Recorder) {
	t.Helper()
	rec := &Recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		response, ok := routes[req.Method+" "+req.URL.EscapedPath()]
		if !ok {
			t.Errorf("unexpected request %s %s", req.Method, req.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if response.ContentType != "" {
			w.Header().Set("Content-Type", response.ContentType)
		}
		w.WriteHeader(response.Status)
		_, _ = io.WriteString(w, response.Body)
	}))
	t.Cleanup(server.Close)
	return server.URL, rec
}
