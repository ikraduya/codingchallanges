package comparisonutils

import "unicode"

type RuneComparisonFunc func(a rune, b rune) bool

func AreRunesCaseSensitiveEqual(a rune, b rune) bool {
	return a == b
}
func AreRunesCaseInsensitiveEqual(a rune, b rune) bool {
	return unicode.ToLower(a) == unicode.ToLower(b)
}
