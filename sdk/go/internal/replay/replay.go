// Copyright 2026, Pulumi Corporation.  All rights reserved.

// Package replay records the HTTP exchanges of a test run against Pulumi Cloud
// and plays them back, so the same test can run with a live backend or from
// its recording.
package replay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type Exchange struct {
	Method              string     `json:"method"`
	Path                string     `json:"path"`
	Query               url.Values `json:"query,omitempty"`
	RequestContentType  string     `json:"requestContentType,omitempty"`
	RequestBody         string     `json:"requestBody,omitempty"`
	Status              int        `json:"status"`
	ResponseContentType string     `json:"responseContentType,omitempty"`
	ResponseBody        string     `json:"responseBody,omitempty"`
}

// Recording is one test run: the inputs the test chose and every exchange it
// made, in order.
type Recording struct {
	Backend   string            `json:"backend"`
	Org       string            `json:"org"`
	Names     map[string]string `json:"names"`
	Exchanges []Exchange        `json:"exchanges"`
}

func Load(path string) (*Recording, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var recording Recording
	if err := json.Unmarshal(content, &recording); err != nil {
		return nil, fmt.Errorf("parsing recording %s: %w", path, err)
	}
	return &recording, nil
}

func (r *Recording) Save(path string) error {
	content, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0o644)
}

// Recorder forwards requests to the next transport and keeps every exchange.
// Request headers are not kept, so the access token never reaches disk.
type Recorder struct {
	next http.RoundTripper

	mu        sync.Mutex
	exchanges []Exchange
}

func NewRecorder(next http.RoundTripper) *Recorder {
	return &Recorder{next: next}
}

func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	requestBody, err := readAndRestoreRequestBody(req)
	if err != nil {
		return nil, err
	}
	resp, err := r.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(responseBody))

	r.mu.Lock()
	defer r.mu.Unlock()
	r.exchanges = append(r.exchanges, Exchange{
		Method:              req.Method,
		Path:                req.URL.EscapedPath(),
		Query:               req.URL.Query(),
		RequestContentType:  req.Header.Get("Content-Type"),
		RequestBody:         string(requestBody),
		Status:              resp.StatusCode,
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseBody:        string(responseBody),
	})
	return resp, nil
}

func (r *Recorder) Exchanges() []Exchange {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Exchange(nil), r.exchanges...)
}

// Player answers requests from a recording in order and fails a request that
// differs from the recorded one.
type Player struct {
	mu        sync.Mutex
	exchanges []Exchange
	position  int
}

func NewPlayer(recording *Recording) *Player {
	return &Player{exchanges: recording.Exchanges}
}

func (p *Player) RoundTrip(req *http.Request) (*http.Response, error) {
	requestBody, err := readAndRestoreRequestBody(req)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.position >= len(p.exchanges) {
		return nil, fmt.Errorf("replay: no recorded exchange left for %s %s", req.Method, req.URL.EscapedPath())
	}
	expected := p.exchanges[p.position]
	p.position++

	if err := matchRequest(expected, req, requestBody); err != nil {
		return nil, fmt.Errorf("replay: exchange %d: %w", p.position, err)
	}
	header := http.Header{}
	if expected.ResponseContentType != "" {
		header.Set("Content-Type", expected.ResponseContentType)
	}
	return &http.Response{
		StatusCode:    expected.Status,
		Status:        fmt.Sprintf("%d %s", expected.Status, http.StatusText(expected.Status)),
		Header:        header,
		Body:          io.NopCloser(strings.NewReader(expected.ResponseBody)),
		ContentLength: int64(len(expected.ResponseBody)),
		Request:       req,
	}, nil
}

// Remaining reports how many recorded exchanges the test did not make.
func (p *Player) Remaining() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.exchanges) - p.position
}

func matchRequest(expected Exchange, req *http.Request, body []byte) error {
	if req.Method != expected.Method || path.Clean(req.URL.EscapedPath()) != path.Clean(expected.Path) {
		return fmt.Errorf("expected %s %s, got %s %s", expected.Method, expected.Path, req.Method, req.URL.EscapedPath())
	}
	if !reflect.DeepEqual(nonEmptyQuery(req.URL.Query()), nonEmptyQuery(expected.Query)) {
		return fmt.Errorf("%s %s: expected query %v, got %v", req.Method, expected.Path, expected.Query, req.URL.Query())
	}
	if !bodiesMatch(expected.RequestContentType, expected.RequestBody, string(body)) {
		return fmt.Errorf("%s %s: expected body %q, got %q", req.Method, expected.Path, expected.RequestBody, string(body))
	}
	return nil
}

func nonEmptyQuery(query url.Values) url.Values {
	if len(query) == 0 {
		return nil
	}
	return query
}

func bodiesMatch(contentType, expected, got string) bool {
	if !strings.Contains(contentType, "json") {
		return expected == got
	}
	var expectedValue, gotValue any
	if json.Unmarshal([]byte(expected), &expectedValue) != nil || json.Unmarshal([]byte(got), &gotValue) != nil {
		return expected == got
	}
	return reflect.DeepEqual(withoutZeroValues(expectedValue), withoutZeroValues(gotValue))
}

func withoutZeroValues(value any) any {
	switch value := value.(type) {
	case map[string]any:
		pruned := map[string]any{}
		for key, member := range value {
			if member := withoutZeroValues(member); member != nil {
				pruned[key] = member
			}
		}
		if len(pruned) == 0 {
			return nil
		}
		return pruned
	case []any:
		if len(value) == 0 {
			return nil
		}
		pruned := make([]any, len(value))
		for i, element := range value {
			pruned[i] = withoutZeroValues(element)
		}
		return pruned
	case string:
		if value == "" {
			return nil
		}
	case float64:
		if value == 0 {
			return nil
		}
	case bool:
		if !value {
			return nil
		}
	case nil:
		return nil
	}
	return value
}

func readAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
