package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

type fakeTool struct {
	name string
	desc string
	out  string
}

func (f *fakeTool) Name() string                { return f.name }
func (f *fakeTool) Description() string         { return f.desc }
func (f *fakeTool) Parameters() json.RawMessage { return json.RawMessage(`{}`) }
func (f *fakeTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return f.out, nil
}

func TestRegistry(t *testing.T) {
	t.Run("should register a tool and retrieve it by name", func(t *testing.T) {
		r := tools.NewRegistry()
		err := r.Register(&fakeTool{name: "t1", out: "ok"})
		require.NoError(t, err)

		got, err := r.Get("t1")
		require.NoError(t, err)
		require.Equal(t, "t1", got.Name())
	})

	t.Run("should return ErrDuplicateTool when registering same name twice", func(t *testing.T) {
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "t1"}))
		err := r.Register(&fakeTool{name: "t1"})
		require.ErrorIs(t, err, tools.ErrDuplicateTool)
	})

	t.Run("should return ErrToolNotFound when getting unknown name", func(t *testing.T) {
		r := tools.NewRegistry()
		_, err := r.Get("missing")
		require.ErrorIs(t, err, tools.ErrToolNotFound)
	})

	t.Run("should list all registered tool names", func(t *testing.T) {
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{name: "t1"}))
		require.NoError(t, r.Register(&fakeTool{name: "t2"}))

		names := r.Names()
		require.ElementsMatch(t, []string{"t1", "t2"}, names)
	})

	t.Run("should convert tools to provider definitions", func(t *testing.T) {
		r := tools.NewRegistry()
		require.NoError(t, r.Register(&fakeTool{
			name: "read_file",
			desc: "reads a file",
		}))

		defs := r.Definitions()
		require.Len(t, defs, 1)
		require.Equal(t, "function", defs[0].Type)
		require.Equal(t, "read_file", defs[0].Function.Name)
		require.Equal(t, "reads a file", defs[0].Function.Description)
	})
}
