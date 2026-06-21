package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestFetchURL(t *testing.T) {
	t.Run("should return content of URL from server", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# Page Title\nPage content here"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewFetchURL()
		tool.BaseURL = server.URL + "/"

		got, err := tool.Execute(context.Background(), []byte(`{"url":"https://example.com/article"}`))
		require.NoError(t, err)
		require.Contains(t, got, "Page Title")
		require.Contains(t, got, "Page content")
		require.Contains(t, capturedPath, "example.com")
	})

	t.Run("should return error on 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		t.Cleanup(server.Close)

		tool := tools.NewFetchURL()
		tool.BaseURL = server.URL + "/"

		_, err := tool.Execute(context.Background(), []byte(`{"url":"https://example.com/missing"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "404")
	})

	t.Run("should return error when url is empty", func(t *testing.T) {
		tool := tools.NewFetchURL()
		_, err := tool.Execute(context.Background(), []byte(`{"url":""}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "url is required")
	})

	t.Run("should return error on parse failure", func(t *testing.T) {
		tool := tools.NewFetchURL()
		_, err := tool.Execute(context.Background(), []byte(`{not valid json`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "fetch_url")
	})
}
