package names

import (
	"strings"
	"unicode"
)

func Normalize(value string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	for i, part := range parts {
		var b strings.Builder
		for _, r := range part {
			if unicode.IsLetter(r) || unicode.IsMark(r) || r == '-' || r == '\'' {
				b.WriteRune(r)
			}
		}
		parts[i] = b.String()
	}
	return strings.Join(parts, " ")
}
