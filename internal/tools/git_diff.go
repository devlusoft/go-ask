package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type GitDiff struct{}

func NewGitDiff() Tool { return &GitDiff{} }

func (t *GitDiff) Name() string { return "git_diff" }

func (t *GitDiff) Description() string {
	return "Shows git diffs. Use staged=true for staged changes, or pass a commit to diff against it."
}

func (t *GitDiff) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "staged":  {"type": "boolean", "description": "Show staged changes instead of unstaged. Defaults to false."},                                                                                                   
            "commit":  {"type": "string", "description": "Commit to diff against (e.g., 'HEAD~1', a hash, or a branch). Optional."},                                                                                        
            "path":    {"type": "string", "description": "Path to filter by. Optional."}                                                                                                                                    
         }                                                                                                                                                                                                                  
      }`)
}

type gitDiffArgs struct {
	Staged bool   `json:"staged"`
	Commit string `json:"commit"`
	Path   string `json:"path"`
}

func (t *GitDiff) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a gitDiffArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("git_diff: parsing args: %w", err)
	}

	cmdArgs := []string{"diff"}
	if a.Staged {
		cmdArgs = append(cmdArgs, "--staged")
	}
	if a.Commit != "" {
		cmdArgs = append(cmdArgs, a.Commit)
	}
	if a.Path != "" {
		cmdArgs = append(cmdArgs, "--", a.Path)
	}

	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git_diff: %w (output: %s)", err, string(out))
	}
	return truncate(string(out), len(out)), nil
}
