package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ListFiles struct{}

func NewListFiles() Tool {
	return &ListFiles{}
}

func (t *ListFiles) Name() string { return "list_files" }

func (t *ListFiles) Description() string {
	return "Lists files in a directory. Set recursive=true to include subdirectories."
}

func (t *ListFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "path":      {"type": "string", "description": "Directory path, relative to CWD. Defaults to current directory."},                                                                                              
            "recursive": {"type": "boolean", "description": "Whether to descend into subdirectories. Defaults to false."}                                                                                                   
         }                                                                                                                                                                                                                  
      }`)
}

type listFilesArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

func (t *ListFiles) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a listFilesArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("list_files: parsing args: %w", err)
	}
	if a.Path == "" {
		a.Path = "."
	}

	var paths []string
	if a.Recursive {
		err := filepath.Walk(a.Path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			paths = append(paths, p)
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("list_files: walking %q: %w", a.Path, err)
		}
	} else {
		entries, err := os.ReadDir(a.Path)
		if err != nil {
			return "", fmt.Errorf("list_files: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			paths = append(paths, filepath.Join(a.Path, e.Name()))
		}
	}

	sort.Strings(paths)
	return truncate(strings.Join(paths, "\n"), -1), nil
}
