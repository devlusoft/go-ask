package render

import (
	"regexp"
	"strings"
)

var thinkingPattern = regexp.MustCompile(`(?s)<think>.*?</think>\s*`)

func StripThinking(input string) string {
	return strings.TrimSpace(thinkingPattern.ReplaceAllString(input, ""))
}
