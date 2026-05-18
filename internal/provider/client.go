package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type clientConfig struct {
	Endpoint              string
	Realm                 string
	Token                 string
	ClientID              string
	ClientSecret          string
	OAuthTokenURL         string
	OAuthScope            string
	InsecureSkipTLSVerify bool
}

type Client struct {
	endpoint      string
	realm         string
	token         string
	clientID      string
	clientSecret  string
	oauthTokenURL string
	oauthScope    string
	httpClient    *http.Client
	tokenMu       sync.Mutex
	UserAgent     string
}

type response struct {
	StatusCode int
	Body       string
}

func newClient(cfg clientConfig) (*Client, error) {
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	parsedEndpoint, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint %q: %w", cfg.Endpoint, err)
	}
	if parsedEndpoint.Scheme != "http" && parsedEndpoint.Scheme != "https" {
		return nil, fmt.Errorf("invalid endpoint %q: scheme must be http or https", cfg.Endpoint)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureSkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return &Client{
		endpoint:      endpoint,
		realm:         cfg.Realm,
		token:         cfg.Token,
		clientID:      cfg.ClientID,
		clientSecret:  cfg.ClientSecret,
		oauthTokenURL: cfg.OAuthTokenURL,
		oauthScope:    cfg.OAuthScope,
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
	}, nil
}

func (c *Client) do(ctx context.Context, method, pathTemplate string, pathParams, queryParams, headers map[string]string, body string) (*response, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}
	expanded, err := expandPath(pathTemplate, pathParams)
	if err != nil {
		return nil, err
	}
	reqURL := c.endpoint + expanded
	if len(queryParams) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), reqURL, reader)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.realm != "" {
		req.Header.Set("Polaris-Realm", c.realm)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &response{StatusCode: resp.StatusCode, Body: string(respBody)}, nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	if c.token != "" {
		return nil
	}
	if c.clientID == "" || c.clientSecret == "" || c.oauthTokenURL == "" {
		return nil
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" {
		return nil
	}

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)
	if c.oauthScope != "" {
		values.Set("scope", c.oauthScope)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.oauthTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("oauth token request failed with HTTP %d: %s", resp.StatusCode, safeHTTPBody(respBody))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return err
	}
	if payload.AccessToken == "" {
		return fmt.Errorf("oauth token response did not contain access_token")
	}
	c.token = payload.AccessToken
	return nil
}
