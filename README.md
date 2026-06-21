# go-ask

One-shot agentic CLI for investigating your repo. Asks questions, get answers.

# What it does

```bash
$ ask "where is the function x?"
[response with code reference]

$ ask "what is the last version of the x library?"
[response with web search results]

$ ask "what is new in this reposity?"
[response with git diff]
```

The LLM investigates your repo using read-only tools, then answers. Works with any OpenAI-compatible provider

# Install

```bash
go install github.com/devlusoft/go-ask@latest
```

# Configure

~/.config/go-ask/config.json
```json
{
  "api_key": "<your-api-key>",
  "base_url": "<your-api-base-url>",
  "model": "<your-model>"
}
```

Or via env vars: ASK_API_KEY, ASK_BASE_URL, ASK_MODEL.
