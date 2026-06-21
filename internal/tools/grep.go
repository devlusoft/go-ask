package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Grep struct{}

func NewGrep() Tool {
	return &Grep{}
}

func (t *Grep) Name() string { return "grep" }

func (t *Grep) Description() string {
	return "Searches files for a regex pattern. Returns matches in 'file:line:content' format. Use the include glob to filter by file type."
}

func (t *Grep) Parameters() json.RawMessage {
	return json.RawMessage(`{                                                                                                                                                                                             
         "type": "object",                                                                                                                                                                                                  
         "properties": {                                                                                                                                                                                                    
            "pattern": {"type": "string", "description": "Regex pattern to search for."},                                                                                                                                   
            "include": {"type": "string", "description": "Glob to filter files (e.g., '*.go'). Defaults to '*'."},                                                                                                          
            "path":   {"type": "string", "description": "Directory to search in. Defaults to current directory."}                                                                                                           
         },                                                                                                                                                                                                                 
         "required": ["pattern"]                                                                                                                                                                                            
      }`)
}

type grepArgs struct {
	Pattern string `json:"pattern"`
	Include string `json:"include"`
	Path    string `json:"path"`
}

func (t *Grep) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a grepArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("grep: parsing args: %w", err)
	}
	if a.Pattern == "" {
		return "", fmt.Errorf("grep: pattern is required")
	}

	re, err := regexp.Compile(a.Pattern)
	if err != nil {
		return "", fmt.Errorf("grep: invalid pattern: %w", err)
	}

	if a.Path == "" {
		a.Path = "."
	}
	if a.Include == "" {
		a.Include = "*"
	}

	var matches []string
	err = filepath.Walk(a.Path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		matched, err := filepath.Match(a.Include, filepath.Base(p))
		if err != nil || !matched {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer func() {
			_ = f.Close()
		}()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			if re.MatchString(scanner.Text()) {
				matches = append(matches, fmt.Sprintf("%s:%d:%s", p, lineNum, scanner.Text()))
			}
		}
		return nil
	})

	var sb strings.Builder
	for _, m := range matches {
		sb.WriteString(m)
		sb.WriteByte('\n')
	}
	return truncate(sb.String(), -1), nil
}
