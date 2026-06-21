package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FetchURL struct {
	HTTP    *http.Client
	BaseURL string
}

func NewFetchURL() *FetchURL {
	return &FetchURL{
		HTTP:    http.DefaultClient,
		BaseURL: "https://r.jina.ai/",
	}
}

func (t *FetchURL) Name() string { return "fetch_url" }

func (t *FetchURL) Description() string {
	return "Fetches a URL and returns its content as markdown."
}

func (t *FetchURL) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "url": {"type": "string", "description": "URL to fetch."}                                                                                                                                                       
         },                                                                                                                                                                                                                 
         "required": ["url"]                                                                                                                                                                                                
      }`)
}

type fetchURLArgs struct {
	URL string `json:"url"`
}

func (t *FetchURL) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a fetchURLArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("fetch_url: parsing args: %w", err)
	}
	if a.URL == "" {
		return "", fmt.Errorf("fetch_url: url is required")
	}

	client := t.HTTP
	if client == nil {
		client = http.DefaultClient
	}

	target := t.BaseURL + a.URL
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("fetch_url: building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch_url: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetch_url: reading body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("fetch_url: status %d: %s", resp.StatusCode, string(body))
	}

	return truncate(string(body), len(body)), nil
}
