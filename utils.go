package gwf

import "strings"

func SubStringLast(s, substr string) string {
	index := strings.Index(s, substr)
	if index == -1 {
		return ""
	}
	return s[index+len(substr):]
}
