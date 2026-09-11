// Copyright 2026, Pulumi Corporation.  All rights reserved.

package replay

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func recordOne(t *testing.T, method, target, contentType, body string) Exchange {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"1"}`)
	}))
	t.Cleanup(server.Close)

	recorder := NewRecorder(http.DefaultTransport)
	req, err := http.NewRequest(method, server.URL+target, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "token secret")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := (&http.Client{Transport: recorder}).Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{"id":"1"}`, string(responseBody), "the caller still reads the body the recorder captured")

	exchanges := recorder.Exchanges()
	require.Len(t, exchanges, 1)
	return exchanges[0]
}

func TestRecorderKeepsTheExchangeWithoutHeaders(t *testing.T) {
	exchange := recordOne(t, http.MethodPost, "/api/esc/environments/acme?x=1&x=2", "application/json", `{"name":"prod"}`)

	require.Equal(t, Exchange{
		Method:              http.MethodPost,
		Path:                "/api/esc/environments/acme",
		Query:               url.Values{"x": {"1", "2"}},
		RequestContentType:  "application/json",
		RequestBody:         `{"name":"prod"}`,
		Status:              http.StatusCreated,
		ResponseContentType: "application/json",
		ResponseBody:        `{"id":"1"}`,
	}, exchange)

	recording := Recording{Org: "acme", Exchanges: []Exchange{exchange}}
	file := filepath.Join(t.TempDir(), "r.json")
	require.NoError(t, recording.Save(file))
	loaded, err := Load(file)
	require.NoError(t, err)
	require.Equal(t, recording, *loaded)
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	require.NotContains(t, string(content), "secret")
}

func play(t *testing.T, recorded Exchange, method, target, contentType, body string) (*http.Response, error) {
	t.Helper()
	player := NewPlayer(&Recording{Exchanges: []Exchange{recorded}})
	req, err := http.NewRequest(method, "https://replay.invalid"+target, strings.NewReader(body))
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return (&http.Client{Transport: player}).Do(req)
}

func TestPlayerAnswersAMatchingRequest(t *testing.T) {
	recorded := Exchange{Method: "PATCH", Path: "/api/esc/environments/acme/p/e", RequestContentType: "application/x-yaml", RequestBody: "values: {}\n", Status: 200, ResponseContentType: "application/json", ResponseBody: `{}`}

	resp, err := play(t, recorded, "PATCH", "/api/esc/environments/acme/p/e", "application/x-yaml", "values: {}\n")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{}`, string(body))
}

func TestPlayerToleratesOmittedZeroValuesInJSONBodies(t *testing.T) {
	recorded := Exchange{Method: "POST", Path: "/clone", RequestContentType: "application/json", RequestBody: `{"name":"copy","preserveHistory":false,"version":0}`, Status: 204}

	_, err := play(t, recorded, "POST", "/clone", "application/json", `{"name":"copy"}`)
	require.NoError(t, err)

	_, err = play(t, recorded, "POST", "/clone", "application/json", `{"name":"copy","preserveHistory":true}`)
	require.ErrorContains(t, err, "expected body")
}

func TestPlayerRejectsADifferentRequest(t *testing.T) {
	recorded := Exchange{Method: "GET", Path: "/a", Query: url.Values{"property": {"x"}}, Status: 200}

	_, err := play(t, recorded, "GET", "/b", "", "")
	require.ErrorContains(t, err, "expected GET /a, got GET /b")

	_, err = play(t, recorded, "GET", "/a?property=y", "", "")
	require.ErrorContains(t, err, "expected query")

	_, err = play(t, recorded, "GET", "/a//?property=x", "", "")
	require.NoError(t, err, "duplicate slashes match, as the service accepts them")
}

func TestPlayerReportsUnusedAndExhaustedRecordings(t *testing.T) {
	player := NewPlayer(&Recording{Exchanges: []Exchange{{Method: "GET", Path: "/a", Status: 200}}})
	require.Equal(t, 1, player.Remaining())

	req, err := http.NewRequest("GET", "https://replay.invalid/a", nil)
	require.NoError(t, err)
	client := &http.Client{Transport: player}
	_, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 0, player.Remaining())

	_, err = client.Do(req)
	require.ErrorContains(t, err, "no recorded exchange left")
}
