package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestWebSearch(t *testing.T) {
	t.Run("should return search results from server", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# Result 1\nContent here\n\n# Result 2\nMore content"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewWebSearch("")
		tool.BaseURL = server.URL + "/"

		got, err := tool.Execute(context.Background(), []byte(`{"query":"hello world"}`))
		require.NoError(t, err)
		require.Contains(t, got, "Result 1")
		require.Contains(t, got, "Result 2")
		require.True(t, strings.Contains(capturedPath, "hello"))
	})

	t.Run("should return error on 5xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("upstream error"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewWebSearch("")
		tool.BaseURL = server.URL + "/"

		_, err := tool.Execute(context.Background(), []byte(`{"query":"hello"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "500")
	})

	t.Run("should return error when query is empty", func(t *testing.T) {
		tool := tools.NewWebSearch("")
		_, err := tool.Execute(context.Background(), []byte(`{"query":""}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "query is required")
	})

	t.Run("should set Authorization header when API key is provided", func(t *testing.T) {
		var capturedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewWebSearch("test-api-key")
		tool.BaseURL = server.URL + "/"

		_, err := tool.Execute(context.Background(), []byte(`{"query":"hello"}`))
		require.NoError(t, err)
		require.Equal(t, "Bearer test-api-key", capturedAuth)
	})

	t.Run("should not set Authorization header when API key is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth := r.Header.Get("Authorization"); auth != "" {
				t.Errorf("Authorization should be empty, got %q", auth)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewWebSearch("")
		tool.BaseURL = server.URL + "/"

		_, err := tool.Execute(context.Background(), []byte(`{"query":"hello"}`))
		require.NoError(t, err)
	})
}
