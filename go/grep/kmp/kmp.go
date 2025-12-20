package kmp

import "ccgrep/comparisonutils"

func ComputeLPSArray(pattern []rune, comparisonFunc comparisonutils.RuneComparisonFunc) []int {
	pat_size := len(pattern)

	lps := make([]int, pat_size)

	len := 0
	i := 1

	for i < pat_size {
		if comparisonFunc(pattern[i], pattern[len]) {
			len += 1
			lps[i] = len
			i += 1
		} else {
			if len != 0 {
				//  fall back in the pattern
				len = lps[len-1]
			} else {
				lps[i] = 0
				i += 1
			}
		}
	}

	return lps
}
