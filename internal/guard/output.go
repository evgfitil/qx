package guard

import (
	"regexp"
	"strings"
)

// controlCharsRegex matches control characters except \t (0x09) and \n (0x0a)
var controlCharsRegex = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)

// explanationPrefixes are phrases that indicate explanatory text rather than commands.
var explanationPrefixes = []string{
	"the command",
	"this command",
	"this will",
	"this would",
	"here is",
	"here's",
	"to do this",
	"you can use",
	"you could use",
	"you should",
	"i would",
	"i'd suggest",
	"i suggest",
	"i recommend",
	"note:",
	"explanation:",
	"the following",
	"below is",
}

// SanitizeOutput removes control characters from LLM response
func SanitizeOutput(s string) string {
	return controlCharsRegex.ReplaceAllString(s, "")
}

// IsExplanation checks if the text appears to be an explanation rather than a command.
func IsExplanation(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	lower := strings.ToLower(trimmed)
	for _, prefix := range explanationPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
