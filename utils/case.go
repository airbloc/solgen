package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ToLowerCase(str string, pos int) string {
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

func Decapitalise(str string) string {
	return ToLowerCase(str, 0)
}

func ToUpperCase(str string, pos int) string {
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

func Capitalize(str string) string {
	return ToUpperCase(str, 0)
}

func ToSnakeCase(str string) string {
	var (
		matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
		matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
	)
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToCamelCase(str string) string {
	var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}
