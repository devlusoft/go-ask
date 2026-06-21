package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAICompat struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func New(baseURL, apiKey, model string) *OpenAICompat {
	return &OpenAICompat{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}
}

func (c *OpenAICompat) Chat(ctx context.Context, messages []Message, tools []ToolDefinition) (Message, error) {
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	body, err := json.Marshal(ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return Message{}, fmt.Errorf("openaicompat: marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Message{}, fmt.Errorf("openaicompat: building request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("openaicompat: calling API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Message{}, fmt.Errorf("openaicompat: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Message{}, &Error{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Message:    fmt.Sprintf("openaicompat: API returned status %d", resp.StatusCode),
		}
	}

	var parsed ChatResponse
	if err = json.Unmarshal(respBody, &parsed); err != nil {
		return Message{}, fmt.Errorf("openaicompat: parsing response: %w (body: %s)", err, string(respBody))
	}

	if len(parsed.Choices) == 0 {
		return Message{}, fmt.Errorf("openaicompat: empty choices in response (body: %s)", string(respBody))
	}

	return parsed.Choices[0].Message, nil
}
