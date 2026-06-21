package provider

import "context"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Provider interface {
	Chat(ctx context.Context, msg Message) (string, error)
}

type Error struct {
	StatusCode int    `json:"-"`
	Body       string `json:"-"`
	Message    string `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}
