package proto

import "unicode"

func toLowerCase(str string, pos int) string {
	if pos > len(str) {
		pos = len(str)
	} else if pos == 0 {
		pos = 0
	}

	runes := []rune(str)
	if len(str) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

func toUpperCase(str string, pos int) string {
	if pos > len(str) {
		pos = len(str)
	} else if pos == 0 {
		pos = 0
	}

	runes := []rune(str)
	if len(str) > 0 {
		runes[pos] = unicode.ToUpper(runes[pos])
	}
	return string(runes)
}
