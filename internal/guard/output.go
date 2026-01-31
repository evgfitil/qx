package guard

import "regexp"

// controlCharsRegex matches control characters except \t (0x09) and \n (0x0a)
var controlCharsRegex = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)

// SanitizeOutput removes control characters from LLM response
func SanitizeOutput(s string) string {
	return controlCharsRegex.ReplaceAllString(s, "")
}
