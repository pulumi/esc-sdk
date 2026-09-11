// Copyright 2026, Pulumi Corporation.  All rights reserved.

package esc_sdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pulumi/pulumi-cloud-sdk/go/apiclient"
)

const DefaultPulumiAPIURL = "https://api.pulumi.com"

const (
	defaultUserAgent = "esc-sdk"
	sourceHeader     = "esc-sdk"
)

// Configuration describes how an EscClient reaches Pulumi Cloud.
type Configuration struct {
	// BaseURL is the scheme and host of the Pulumi Cloud API, without a path.
	BaseURL       string
	UserAgent     string
	DefaultHeader map[string]string
	// HTTPClient sends every request. http.DefaultClient is used when nil.
	HTTPClient *http.Client
}

func NewConfiguration() *Configuration {
	return &Configuration{
		BaseURL:       DefaultPulumiAPIURL,
		UserAgent:     defaultUserAgent,
		DefaultHeader: map[string]string{},
	}
}

// AddDefaultHeader sends the header with every request.
func (c *Configuration) AddDefaultHeader(key, value string) {
	if c.DefaultHeader == nil {
		c.DefaultHeader = map[string]string{}
	}
	c.DefaultHeader[key] = value
}

func (c *Configuration) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// EscClient is the hand-written ESC API over the generated Pulumi Cloud client.
type EscClient struct {
	cfg *Configuration
	// Cloud is the generated client every EscClient method calls, configured
	// with this client's authentication and headers. It reaches the whole
	// Pulumi Cloud API, not only ESC.
	Cloud *apiclient.CloudClient
}

func NewClient(cfg *Configuration) *EscClient {
	if cfg == nil {
		cfg = NewConfiguration()
	}
	client := &EscClient{cfg: cfg}
	client.Cloud = &apiclient.CloudClient{BaseURL: cfg.BaseURL, Executor: client.execute}
	return client
}

// NewCustomBackendConfiguration configures a client for a self-hosted Pulumi
// Cloud. Only the scheme and host of customBackendURL are used.
func NewCustomBackendConfiguration(customBackendURL url.URL) (*Configuration, error) {
	if customBackendURL.Scheme == "" || customBackendURL.Host == "" {
		return nil, fmt.Errorf("backend url %q must include a scheme and a host", customBackendURL.String())
	}
	cfg := NewConfiguration()
	cfg.BaseURL = customBackendURL.Scheme + "://" + customBackendURL.Host
	return cfg, nil
}

// NewDefaultClient reads the backend from PULUMI_BACKEND_URL and falls back to
// Pulumi Cloud.
func NewDefaultClient() (*EscClient, error) {
	backendURL := os.Getenv("PULUMI_BACKEND_URL")
	if backendURL == "" {
		backendURL = DefaultPulumiAPIURL
	}
	parsedURL, err := url.Parse(backendURL)
	if err != nil {
		return nil, fmt.Errorf("parsing backend url: %w", err)
	}
	cfg, err := NewCustomBackendConfiguration(*parsedURL)
	if err != nil {
		return nil, err
	}
	return NewClient(cfg), nil
}

type accessTokenKey struct{}

// NewAuthContext returns a context whose requests authenticate with accessToken.
func NewAuthContext(accessToken string) context.Context {
	return WithAccessToken(context.Background(), accessToken)
}

func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

func AccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(accessTokenKey{}).(string)
	return token, ok && token != ""
}

// NewDefaultAuthContext authenticates with PULUMI_ACCESS_TOKEN.
func NewDefaultAuthContext() (context.Context, error) {
	accessToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("no Pulumi Access Token found. Export the PULUMI_ACCESS_TOKEN environment variable")
	}
	return NewAuthContext(accessToken), nil
}

// DefaultLogin builds a client and an authenticated context from the environment.
func DefaultLogin() (context.Context, *EscClient, error) {
	client, err := NewDefaultClient()
	if err != nil {
		return nil, nil, err
	}
	ctx, err := NewDefaultAuthContext()
	if err != nil {
		return nil, nil, err
	}
	return ctx, client, nil
}

func (c *EscClient) execute(req *http.Request) (*http.Response, error) {
	if token, ok := AccessTokenFromContext(req.Context()); ok {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("X-Pulumi-Source", sourceHeader)
	for key, value := range c.cfg.DefaultHeader {
		req.Header.Set(key, value)
	}
	resp, err := c.cfg.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if capture := responseCaptureFromContext(req.Context()); capture != nil {
		if err := capture.observe(resp); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

type responseCapture struct {
	header http.Header
	body   []byte
}

type responseCaptureKey struct{}

func withResponseCapture(ctx context.Context) (context.Context, *responseCapture) {
	capture := &responseCapture{}
	return context.WithValue(ctx, responseCaptureKey{}, capture), capture
}

func responseCaptureFromContext(ctx context.Context) *responseCapture {
	capture, _ := ctx.Value(responseCaptureKey{}).(*responseCapture)
	return capture
}

func (c *responseCapture) observe(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	c.header = resp.Header.Clone()
	c.body = body
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return nil
}
