package match

import (
	"ccgrep/comparisonutils"
	"ccgrep/kmp"

	"unicode"
)

type ExpressionOption struct {
	IsInvertExpression bool
	IsCaseInsensitive  bool
}

type IndexRange struct {
	Start int
	Stop  int
}

// Contract: return nil if no match; return a non-empty slice if matched.
type CheckContainsOperation func(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange

func ContainsExpression(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange {
	indexRanges := make([]IndexRange, 0)
	if len(exp) == 0 {
		indexRanges = append(indexRanges, IndexRange{Start: 0, Stop: len(s)})
		return indexRanges
	}

	runeComparisonFunc := comparisonutils.AreRunesCaseSensitiveEqual
	if expOptions.IsCaseInsensitive {
		runeComparisonFunc = comparisonutils.AreRunesCaseInsensitiveEqual
	}

	// Using knuth morris prat
	lps := kmp.ComputeLPSArray(exp, runeComparisonFunc)

	sLen := len(s)
	patLen := len(exp)

	sIdx := 0
	patIdx := 0
	for sIdx < sLen {
		// if character match, move pointer forward
		if runeComparisonFunc(s[sIdx], exp[patIdx]) {
			sIdx += 1
			patIdx += 1

			// entire pattern is match
			if patIdx == patLen {
				startIdx := sIdx - patIdx
				indexRanges = append(indexRanges, IndexRange{Start: startIdx, Stop: startIdx + patLen})

				patIdx = lps[patIdx-1]
			}
		} else {
			// use lps of previous index
			if patIdx != 0 {
				patIdx = lps[patIdx-1]
			} else {
				sIdx += 1
			}
		}
	}

	if len(indexRanges) > 0 {
		return indexRanges
	} else {
		return nil
	}
}

func ContainsDigit(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange {
	indexRanges := make([]IndexRange, 0)

	indexStart := -1
	indexStop := -1

	for idx, r := range s {
		if unicode.IsDigit(r) {
			if indexStop == -1 {
				indexStart = idx
				indexStop = idx
			} else {
				indexStop = idx
			}
		} else { // not digit
			if indexStart != -1 { // currently in a range, so stop the range and store it
				indexStop = idx
				indexRanges = append(indexRanges, IndexRange{Start: indexStart, Stop: indexStop})

				indexStart = -1
				indexStop = -1
			}
		}
	}
	if indexStart != -1 {
		indexStop += 1
		indexRanges = append(indexRanges, IndexRange{Start: indexStart, Stop: indexStop})
	}

	if len(indexRanges) > 0 {
		return indexRanges
	} else {
		return nil
	}
}

func ContainsWordCharacter(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange {
	indexRanges := make([]IndexRange, 0)

	indexStart := -1
	indexStop := -1

	wordChars := []*unicode.RangeTable{unicode.Digit, unicode.Letter}

	for idx, r := range s {
		if unicode.IsOneOf(wordChars, r) {
			if indexStop == -1 {
				indexStart = idx
				indexStop = idx
			} else {
				indexStop = idx
			}
		} else { // not digit
			if indexStart != -1 { // currently in a range, so stop the range and store it
				indexStop = idx
				indexRanges = append(indexRanges, IndexRange{Start: indexStart, Stop: indexStop})

				indexStart = -1
				indexStop = -1
			}
		}
	}
	if indexStart != -1 {
		indexStop += 1
		indexRanges = append(indexRanges, IndexRange{Start: indexStart, Stop: indexStop})
	}

	if len(indexRanges) > 0 {
		return indexRanges
	} else {
		return nil
	}
}

func ContainsBeginningExpression(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange {
	indexRange := make([]IndexRange, 0, 1)

	runeComparisonFunc := comparisonutils.AreRunesCaseSensitiveEqual
	if expOptions.IsCaseInsensitive {
		runeComparisonFunc = comparisonutils.AreRunesCaseInsensitiveEqual
	}

	expLen := len(exp)
	sLen := len(s)
	if expLen > sLen {
		return nil
	}

	for i := 0; i < expLen; i += 1 {
		if !runeComparisonFunc(s[i], exp[i]) {
			return nil
		}
	}

	indexRange = append(indexRange, IndexRange{Start: 0, Stop: len(exp)})
	return indexRange
}

func trimLineEndingEnd(s []rune) []rune {
	endIdx := len(s)
	if endIdx > 0 && s[endIdx-1] == '\n' {
		endIdx--
		if endIdx > 0 && s[endIdx-1] == '\r' {
			endIdx--
		}
	}

	return s[:endIdx]
}

func ContainsEndingExpression(s []rune, exp []rune, expOptions ExpressionOption) []IndexRange {
	indexRange := make([]IndexRange, 0, 1)

	runeComparisonFunc := comparisonutils.AreRunesCaseSensitiveEqual
	if expOptions.IsCaseInsensitive {
		runeComparisonFunc = comparisonutils.AreRunesCaseInsensitiveEqual
	}

	s = trimLineEndingEnd(s)

	expLen := len(exp)
	sLen := len(s)
	if expLen > sLen {
		return nil
	}

	for i := 0; i < expLen; i += 1 {
		sIdx := sLen - expLen + i
		if !runeComparisonFunc(s[sIdx], exp[i]) {
			return nil
		}
	}

	indexRange = append(indexRange, IndexRange{Start: len(s) - len(exp), Stop: len(s)})
	return indexRange
}
