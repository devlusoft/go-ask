package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devlusoft/go-ask/internal/provider"
	"github.com/devlusoft/go-ask/internal/tools"
)

var (
	ErrMaxIterations = errors.New("agent: max iterations reached")
)

const defaultMaxIterations = 10

const defaultSystemPrompt = `You are go-ask, a one-shot CLI agent. You answer questions about the user's repository using read-only tools.

Current date and time: %s
Working directory: %s

Available tools: read_file, list_files, grep, glob, git_log, git_diff, fetch_url, web_search.

Behavior:
- You give a complete, self-contained answer that fully resolves the user's question in this single response
- You end with the answer, not with a question
- You always assume this is the last interaction, so your response is the final answer
- You answer in the same language the user used
- You use tools (read_file, grep, etc.) whenever they help you answer accurately

Your response is the final answer.`

type Agent struct {
	Provider     provider.Provider
	Registry     *tools.Registry
	MaxIters     int
	SystemPrompt string
	OnToolCall   func(name string, args string)
}

func New(p provider.Provider, r *tools.Registry) *Agent {
	cwd, _ := os.Getwd()
	return &Agent{
		Provider:     p,
		Registry:     r,
		MaxIters:     defaultMaxIterations,
		SystemPrompt: fmt.Sprintf(defaultSystemPrompt, time.Now().Format("2006-01-02 15:04:05"), cwd),
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
				result = fmt.Sprintf("error: %s", tErr)
			}
			messages = append(messages, provider.Message{
				Role:       provider.RoleTool,
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}

	messages = append(messages, provider.Message{
		Role:    provider.RoleUser,
		Content: "You have reached the maximum number of iterations. Give your best answer based on the information you have so far.",
	})

	finalResp, err := a.Provider.Chat(ctx, messages, toolsDef)
	if err != nil {
		return "", fmt.Errorf("agent: final answer call: %w", err)
	}
	return finalResp.Content, nil
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
