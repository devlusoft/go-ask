package provider_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devlusoft/go-ask/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestOpenAICompat_Chat(t *testing.T) {
	t.Run("should return assistant content on 200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{                                                                                                                                                                                     
               "choices": [                                                                                                                                                                                                 
                  {"message": {"role": "assistant", "content": "hello from mock"}}                                                                                                                                          
               ]                                                                                                                                                                                                            
            }`)
		}))
		t.Cleanup(server.Close)

		c := provider.New(server.URL, "test-key", "test-model")
		got, err := c.Chat(context.Background(), provider.Message{
			Role:    provider.RoleUser,
			Content: "hi",
		})
		require.NoError(t, err)
		require.Equal(t, "hello from mock", got)
	})

	t.Run("should POST to {baseURL}/chat/completions", func(t *testing.T) {
		var capturedMethod, capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedMethod = r.Method
			capturedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
		}))
		t.Cleanup(server.Close)

		c := provider.New(server.URL, "test-key", "test-model")
		_, err := c.Chat(context.Background(), provider.Message{
			Role:    provider.RoleUser,
			Content: "hi",
		})
		require.NoError(t, err)
		require.Equal(t, http.MethodPost, capturedMethod)
		require.Equal(t, "/chat/completions", capturedPath)
	})

	t.Run("should set Authorization header with bearer token", func(t *testing.T) {
		var capturedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
		}))
		t.Cleanup(server.Close)

		c := provider.New(server.URL, "test-key", "test-model")
		_, err := c.Chat(context.Background(), provider.Message{
			Role:    provider.RoleUser,
			Content: "hi",
		})
		require.NoError(t, err)
		require.Equal(t, "Bearer test-key", capturedAuth)
	})

	t.Run("should return error on malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `this is not json`)
		}))
		t.Cleanup(server.Close)

		c := provider.New(server.URL, "test-key", "test-model")
		_, err := c.Chat(context.Background(), provider.Message{
			Role:    provider.RoleUser,
			Content: "hi",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "openaicompat")
	})

	t.Run("should return error on 401 with response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":{"message":"invalid api key","type":"invalid_request_error"}}`)
		}))
		t.Cleanup(server.Close)

		c := provider.New(server.URL, "bad-key", "test-model")
		_, err := c.Chat(context.Background(), provider.Message{
			Role:    provider.RoleUser,
			Content: "hi",
		})
		require.Error(t, err)

		var apiErr *provider.Error
		require.True(t, errors.As(err, &apiErr), "error should be unwrappable to *provider.Error")
		require.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
		require.Contains(t, apiErr.Body, "invalid api key")
		require.NotEmpty(t, apiErr.Message)
	})
}
