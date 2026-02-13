package llm

import (
	"regexp"
	"strings"
)

func trimTrailingSpace(b *strings.Builder) {
	s := b.String()
	trimmed := strings.TrimRight(s, " ")
	b.Reset()
	b.WriteString(trimmed)
}

// FormatCommand formats shell command with line breaks at pipe operators.
// Splits at |, &&, || outside of quotes.
func FormatCommand(cmd string) string {
	var result strings.Builder
	inSingle := false
	inDouble := false
	escaped := false
	i := 0

	for i < len(cmd) {
		c := cmd[i]

		if escaped {
			result.WriteByte(c)
			escaped = false
			i++
			continue
		}

		if c == '\\' {
			escaped = true
			result.WriteByte(c)
			i++
			continue
		}

		if c == '\'' && !inDouble {
			inSingle = !inSingle
			result.WriteByte(c)
			i++
			continue
		}

		if c == '"' && !inSingle {
			inDouble = !inDouble
			result.WriteByte(c)
			i++
			continue
		}

		if !inSingle && !inDouble {
			if c == '|' && i+1 < len(cmd) && cmd[i+1] == '|' {
				trimTrailingSpace(&result)
				result.WriteString(" \\\n\t|| ")
				i += 2
				for i < len(cmd) && cmd[i] == ' ' {
					i++
				}
				continue
			}
			if c == '&' && i+1 < len(cmd) && cmd[i+1] == '&' {
				trimTrailingSpace(&result)
				result.WriteString(" \\\n\t&& ")
				i += 2
				for i < len(cmd) && cmd[i] == ' ' {
					i++
				}
				continue
			}
			if c == '|' && (i+1 >= len(cmd) || cmd[i+1] != '|') {
				trimTrailingSpace(&result)
				result.WriteString(" \\\n\t| ")
				i++
				for i < len(cmd) && cmd[i] == ' ' {
					i++
				}
				continue
			}
		}

		result.WriteByte(c)
		i++
	}

	return strings.TrimSpace(result.String())
}

var continuationRe = regexp.MustCompile(`[ \t]*\\\n[\t ]*`)

// UnformatCommand reverses FormatCommand by joining line continuations
// back into a single-line command. Sequences of \<newline><whitespace>
// are collapsed into a single space.
func UnformatCommand(cmd string) string {
	if !strings.Contains(cmd, "\\\n") {
		return cmd
	}
	return strings.TrimSpace(continuationRe.ReplaceAllString(cmd, " "))
}
