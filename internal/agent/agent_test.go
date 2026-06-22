package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/devlusoft/go-ask/internal/agent"
	"github.com/devlusoft/go-ask/internal/provider"
	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	responses []provider.Message
	callIdx   int
}

func (f *fakeProvider) Chat(_ context.Context, _ []provider.Message, _ []provider.ToolDefinition) (provider.Message, error) {
	if f.callIdx >= len(f.responses) {
		return provider.Message{}, errors.New("fakeProvider: no more scripted responses")
	}
	resp := f.responses[f.callIdx]
	f.callIdx++
	return resp, nil
}

type fakeTool struct {
	name string
	out  string
	err  error
}

func (f *fakeTool) Name() string                { return f.name }
func (f *fakeTool) Description() string         { return "fake tool" }
func (f *fakeTool) Parameters() json.RawMessage { return json.RawMessage(`{}`) }
func (f *fakeTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.out, nil
}

func TestAgent_Run(t *testing.T) {
	t.Run("should return final content when LLM responds without tool calls", func(t *testing.T) {
		p := &fakeProvider{responses: []provider.Message{
			{Role: provider.RoleAssistant, Content: "the answer"},
		}}
		r := tools.NewRegistry()
		a := agent.New(p, r)

		got, err := a.Run(context.Background(), "what is X?")
		require.NoError(t, err)
		require.Equal(t, "the answer", got)
	})

	t.Run("should execute a single tool and continue the loop", func(t *testing.T) {
		p := &fakeProvider{responses: []provider.Message{
			{
				Role:    provider.RoleAssistant,
				Content: "",
				ToolCalls: []provider.ToolCall{{
					ID:   "call_1",
					Type: "function",
					Function: provider.ToolCallFunc{
						Name:      "read_file",
						Arguments: `{"path":"main.go"}`,
					},
				}},
			},
			{Role: provider.RoleAssistant, Content: "based on the file, the answer is 42"},
		}}
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "read_file", out: "package main"}))
		a := agent.New(p, r)

		got, err := a.Run(context.Background(), "what's in main.go?")
		require.NoError(t, err)
		require.Equal(t, "based on the file, the answer is 42", got)
	})

	t.Run("should handle multiple tool calls in a single response", func(t *testing.T) {
		p := &fakeProvider{responses: []provider.Message{
			{
				Role: provider.RoleAssistant,
				ToolCalls: []provider.ToolCall{
					{ID: "c1", Type: "function", Function: provider.ToolCallFunc{Name: "t1", Arguments: `{}`}},
					{ID: "c2", Type: "function", Function: provider.ToolCallFunc{Name: "t2", Arguments: `{}`}},
				},
			},
			{Role: provider.RoleAssistant, Content: "both done"},
		}}
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "t1", out: "r1"}))
		require.NoError(t, r.Register(&fakeTool{name: "t2", out: "r2"}))
		a := agent.New(p, r)

		got, err := a.Run(context.Background(), "do both")
		require.NoError(t, err)
		require.Equal(t, "both done", got)
		require.Equal(t, 2, p.callIdx)
	})

	t.Run("should give LLM a final chance when max iterations reached", func(t *testing.T) {
		toolCall := provider.Message{
			Role: provider.RoleAssistant,
			ToolCalls: []provider.ToolCall{{
				ID: "c1", Type: "function",
				Function: provider.ToolCallFunc{Name: "t1", Arguments: `{}`},
			}},
		}
		responses := make([]provider.Message, 3)
		for i := range responses {
			responses[i] = toolCall
		}
		responses = append(responses, provider.Message{
			Role:    provider.RoleAssistant,
			Content: "best partial answer based on what I found",
		})

		p := &fakeProvider{responses: responses}
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "t1", out: "ok"}))
		a := agent.New(p, r)
		a.MaxIters = 3

		got, err := a.Run(context.Background(), "loop forever")
		require.NoError(t, err)
		require.Equal(t, "best partial answer based on what I found", got)
		require.Equal(t, 4, p.callIdx)
	})

	t.Run("should surface tool errors to the LLM as tool results", func(t *testing.T) {
		p := &fakeProvider{responses: []provider.Message{
			{
				Role: provider.RoleAssistant,
				ToolCalls: []provider.ToolCall{{
					ID: "c1", Type: "function",
					Function: provider.ToolCallFunc{Name: "broken", Arguments: `{}`},
				}},
			},
			{Role: provider.RoleAssistant, Content: "I couldn't use that tool, sorry"},
		}}
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "broken", err: errors.New("boom")}))
		a := agent.New(p, r)

		got, err := a.Run(context.Background(), "use broken")
		require.NoError(t, err)
		require.Equal(t, "I couldn't use that tool, sorry", got)
	})
}
