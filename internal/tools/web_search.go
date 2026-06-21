package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type WebSearch struct {
	HTTP    *http.Client
	BaseURL string
}

func NewWebSearch() *WebSearch {
	return &WebSearch{
		HTTP:    http.DefaultClient,
		BaseURL: "https://s.jina.ai/",
	}
}

func (t *WebSearch) Name() string { return "web_search" }

func (t *WebSearch) Description() string {
	return "Searches the web for the given query. Returns results in markdown format."
}

func (t *WebSearch) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "query": {"type": "string", "description": "Search query."},                                                                                                                                                    
            "n":     {"type": "integer", "description": "Number of results. Defaults to 5."}                                                                                                                                
         },                                                                                                                                                                                                                 
         "required": ["query"]                                                                                                                                                                                              
      }`)
}

type webSearchArgs struct {
	Query string `json:"query"`
	N     int    `json:"n"`
}

func (t *WebSearch) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a webSearchArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("web_search: parsing args: %w", err)
	}
	if a.Query == "" {
		return "", fmt.Errorf("web_search: query is required")
	}

	client := t.HTTP
	if client == nil {
		client = http.DefaultClient
	}

	target := t.BaseURL + url.QueryEscape(a.Query)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("web_search: building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_search: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("web_search: reading body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("web_search: status %d: %s", resp.StatusCode, string(body))
	}

	return truncate(string(body), len(body)), nil
}
