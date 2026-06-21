package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devlusoft/go-ask/internal/provider"
	"github.com/devlusoft/go-ask/internal/tools"
)

var (
	ErrMaxIterations = errors.New("agent: max iterations reached")
)

const defaultMaxIterations = 10

const defaultSystemPrompt = "You are a helpful assistant with access to tools. Use them when needed to answer the user's question."

type Agent struct {
	Provider     provider.Provider
	Registry     *tools.Registry
	MaxIters     int
	SystemPrompt string
	OnToolCall   func(name string, args string)
}

func New(p provider.Provider, r *tools.Registry) *Agent {
	return &Agent{
		Provider:     p,
		Registry:     r,
		MaxIters:     defaultMaxIterations,
		SystemPrompt: defaultSystemPrompt,
	}
}

func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	maxIters := a.MaxIters
	if maxIters <= 0 {
		maxIters = defaultMaxIterations
	}
	systemPrompt := a.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: systemPrompt},
		{Role: provider.RoleUser, Content: prompt},
	}

	toolsDef := a.Registry.Definitions()

	for i := 0; i < maxIters; i++ {
		resp, err := a.Provider.Chat(ctx, messages, toolsDef)
		if err != nil {
			return "", fmt.Errorf("agent: provider call: %w", err)
		}

		messages = append(messages, resp)

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		for _, tc := range resp.ToolCalls {
			if a.OnToolCall != nil {
				a.OnToolCall(tc.Function.Name, tc.Function.Arguments)
			}
			result, tErr := a.executeTool(ctx, tc)
			if tErr != nil {
				return "", fmt.Errorf("agent: tool execution: %w", tErr)
			}
			messages = append(messages, provider.Message{
				Role:       provider.RoleTool,
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}

	return "", fmt.Errorf("%w (%d)", ErrMaxIterations, maxIters)
}

func (a *Agent) executeTool(ctx context.Context, tc provider.ToolCall) (string, error) {
	tool, err := a.Registry.Get(tc.Function.Name)
	if err != nil {
		return "", fmt.Errorf("agent: tool %q: %w", tc.Function.Name, err)
	}

	var args json.RawMessage
	if tc.Function.Arguments != "" {
		if err = json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("agent: parsing args for %q: %w", tc.Function.Name, err)
		}
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		return "", fmt.Errorf("agent: executing %q: %w", tc.Function.Name, err)
	}
	return result, nil
}
