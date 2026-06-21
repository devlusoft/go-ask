package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type GitLog struct{}

func NewGitLog() Tool {
	return &GitLog{}
}

func (t *GitLog) Name() string { return "git_log" }

func (t *GitLog) Description() string {
	return "Shows recent git commit history. Use the 'path' argument to filter by file or directory."
}

func (t *GitLog) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "path":    {"type": "string", "description": "Path to filter commits by. Optional."},                                                                                                                           
            "n":       {"type": "integer", "description": "Number of commits to show. Defaults to 20."},                                                                                                                    
            "oneline": {"type": "boolean", "description": "Use short format (hash + subject). Defaults to true."}                                                                                                           
         }                                                                                                                                                                                                                  
      }`)
}

type gitLogArgs struct {
	Path    string `json:"path"`
	N       int    `json:"n"`
	Oneline bool   `json:"oneline"`
}

func (t *GitLog) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a gitLogArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("git_log: parsing args: %w", err)
	}
	if a.N == 0 {
		a.N = 20
	}

	cmdArgs := []string{"log", "-n", strconv.Itoa(a.N)}
	if a.Oneline {
		cmdArgs = append(cmdArgs, "--pretty=format:%h %s")
	}
	if a.Path != "" {
		cmdArgs = append(cmdArgs, "--", a.Path)
	}

	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git_log: %w (output: %s)", err, string(out))
	}
	return truncate(string(out), len(out)), nil
}
