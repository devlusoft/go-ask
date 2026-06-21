package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type Glob struct{}

func NewGlob() Tool {
	return &Glob{}
}

func (t *Glob) Name() string { return "glob" }

func (t *Glob) Description() string {
	return "Finds files matching a glob pattern. Supports ** for recursive matching (e.g., '**/*.go')."
}

func (t *Glob) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "pattern": {"type": "string", "description": "Glob pattern. Supports ** for recursive matching."}                                                                                                               
         },                                                                                                                                                                                                                 
         "required": ["pattern"]                                                                                                                                                                                            
      }`)
}

type globArgs struct {
	Pattern string `json:"pattern"`
}

func (t *Glob) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a globArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("glob: parsing args: %w", err)
	}
	if a.Pattern == "" {
		return "", fmt.Errorf("glob: pattern is required")
	}

	matches, err := doublestar.FilepathGlob(a.Pattern)
	if err != nil {
		return "", fmt.Errorf("glob: %w", err)
	}
	sort.Strings(matches)
	return truncate(strings.Join(matches, "\n"), -1), nil
}
