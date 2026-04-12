package render

import "strings"

func parseFence(line string) (bool, byte, int, string) {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 3 {
		return false, 0, 0, ""
	}
	char := trimmed[0]
	if char != '`' && char != '~' {
		return false, 0, 0, ""
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == char {
		count++
	}
	if count < 3 {
		return false, 0, 0, ""
	}
	return true, char, count, strings.TrimSpace(trimmed[count:])
}

func isFenceClose(line string, fenceChar byte, fenceLen int) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < fenceLen {
		return false
	}
	for index := 0; index < fenceLen; index++ {
		if trimmed[index] != fenceChar {
			return false
		}
	}
	return strings.TrimSpace(trimmed[fenceLen:]) == ""
}
