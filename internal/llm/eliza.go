package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

// ElizaProvider implements Provider interface for Eliza API.
// handles OAuth authorization and Eliza-specific response format unwrapping.
type ElizaProvider struct {
	baseProvider
}

// elizaTransport handles OAuth auth and response unwrapping
type elizaTransport struct {
	apiKey    string
	transport http.RoundTripper
}

func (t *elizaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "OAuth "+t.apiKey)

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	return unwrapElizaResponse(resp)
}

// unwrapElizaResponse extracts the actual OpenAI-compatible response from Eliza wrapper.
func unwrapElizaResponse(resp *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return resp, err
	}

	var wrapper struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	resp.Body = io.NopCloser(bytes.NewReader(wrapper.Response))
	resp.ContentLength = int64(len(wrapper.Response))

	return resp, nil
}

// newElizaProvider creates a new Yandex Eliza provider.
// Uses OAuth authorization and handles Eliza-specific response format.
func newElizaProvider(cfg Config) (*ElizaProvider, error) {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	config.HTTPClient = &http.Client{
		Transport: &elizaTransport{
			apiKey:    cfg.APIKey,
			transport: http.DefaultTransport,
		},
	}

	return &ElizaProvider{
		baseProvider: baseProvider{
			client: openai.NewClientWithConfig(config),
			model:  cfg.Model,
		},
	}, nil
}
