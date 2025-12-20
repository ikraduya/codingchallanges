package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/fatih/color"

	"ccgrep/comparisonutils"
	"ccgrep/kmp"
)

type Args struct {
	Expression string
	Filepaths  []string

	IsRecurse          bool
	IsInvertExpression bool
	IsCaseInsensitive  bool

	ExeName        string
	UseStdInStream bool
}

func printHelp() {
	fmt.Println("Usage: ./ccgrep [OPTION]... EXPRESSION [FILE]...")
	fmt.Println("OPTION: \n\t'-r' Recurse the directory tree")
	fmt.Println("\t'-v' Inverse the match expression")
	fmt.Println("\t'-i' Case insensitive match")
}

func (args *Args) Parse() (isValid bool) {
	args.ExeName = os.Args[0]
	args.ExeName = strings.TrimPrefix(args.ExeName, "./")

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		printHelp()
		return false
	}

	args.Filepaths = make([]string, 0, len(argsWithoutProg)-1)

	positionIndex := 0
	for _, value := range argsWithoutProg {
		if value == args.ExeName {
			continue
		}

		switch value {
		case "-r":
			args.IsRecurse = true
		case "-v":
			args.IsInvertExpression = true
		case "-i":
			args.IsCaseInsensitive = true
		default:
			if positionIndex == 0 {
				args.Expression = value
			} else {
				args.Filepaths = append(args.Filepaths, value)
			}
			positionIndex += 1
		}
	}

	switch positionIndex {
	case 0:
		fmt.Println("EXPRESSION is required")
		printHelp()
		return false
	case 1:
		args.UseStdInStream = true
	} // positionIndex > 1

	return true
}

type ExpressionOption struct {
	IsInvertExpression bool
	IsCaseInsensitive  bool
}

func isFileExistsAndRegular(filepath string) bool {
	info, err := os.Stat(filepath)
	if err == nil {
		return info.Mode().IsRegular()
	}

	return false
}

func isFolderExists(filepath string) (bool, error) {
	info, err := os.Stat(filepath)
	if err == nil {
		return info.Mode().IsDir(), nil
	}

	return false, err
}

type IndexRange struct {
	Start int
	Stop  int
}

type CheckContainsOperation func(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange)

func containsExpression(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange) {
	indexRanges := make([]IndexRange, 0)
	if len(exp) == 0 {
		indexRanges = append(indexRanges, IndexRange{Start: 0, Stop: len(s)})
		return true, indexRanges
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
		return true, indexRanges
	} else {
		return false, nil
	}
}

func containsDigit(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange) {
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

	return len(indexRanges) > 0, indexRanges
}

func containsWordCharacter(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange) {
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

	return len(indexRanges) > 0, indexRanges
}

func containsBeginningExpression(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange) {
	indexRange := make([]IndexRange, 0, 1)

	runeComparisonFunc := comparisonutils.AreRunesCaseSensitiveEqual
	if expOptions.IsCaseInsensitive {
		runeComparisonFunc = comparisonutils.AreRunesCaseInsensitiveEqual
	}

	expLen := len(exp)
	sLen := len(s)
	if expLen > sLen {
		return false, nil
	}

	for i := 0; i < expLen; i += 1 {
		if !runeComparisonFunc(s[i], exp[i]) {
			return false, nil
		}
	}

	indexRange = append(indexRange, IndexRange{Start: 0, Stop: len(exp)})
	return true, indexRange
}

func containsEndingExpression(s []rune, exp []rune, expOptions ExpressionOption) (bool, []IndexRange) {
	indexRange := make([]IndexRange, 0, 1)

	runeComparisonFunc := comparisonutils.AreRunesCaseSensitiveEqual
	if expOptions.IsCaseInsensitive {
		runeComparisonFunc = comparisonutils.AreRunesCaseInsensitiveEqual
	}

	s = []rune(strings.TrimSpace(string(s)))

	expLen := len(exp)
	sLen := len(s)
	if expLen > sLen {
		return false, nil
	}

	for i := 0; i < expLen; i += 1 {
		sIdx := sLen - expLen + i
		if !runeComparisonFunc(s[sIdx], exp[i]) {
			return false, nil
		}
	}

	indexRange = append(indexRange, IndexRange{Start: len(s) - len(exp), Stop: len(s)})
	return true, indexRange
}

var purpleTextColor func(a ...interface{}) string = color.New(color.FgHiMagenta).SprintFunc()
var redTextColor func(a ...interface{}) string = color.New(color.FgHiRed, color.Bold).SprintFunc()
var defaultTextColor = func(a ...interface{}) string { return fmt.Sprint(a...) }

func scanExpressionPattern(scanner *bufio.Scanner, checkContainsFunc CheckContainsOperation, exp string, expOptions ExpressionOption, printPrefix string, lastExtraNewLine bool) int {
	usePrintPrefix := printPrefix != ""

	var matchedTextColor func(a ...interface{}) string = redTextColor
	if len(exp) == 0 {
		matchedTextColor = defaultTextColor
	}
	expRunes := []rune(exp)

	printedCount := 0
	for scanner.Scan() {
		lineString := scanner.Text()
		line := []rune(lineString)

		isMatched, indexRanges := checkContainsFunc(line, expRunes, expOptions)
		if isMatched && len(indexRanges) == 0 {
			// shouldn't happen
			continue
		}

		if !expOptions.IsInvertExpression && isMatched {
			printedCount += 1

			if usePrintPrefix {
				fmt.Printf("%s:", purpleTextColor(printPrefix))
			}

			if indexRanges[0].Start > 0 {
				fmt.Printf("%s", string(line[:indexRanges[0].Start]))
			}
			for i := 0; i < len(indexRanges)-1; i++ {
				fmt.Printf("%s", matchedTextColor(string(line[indexRanges[i].Start:indexRanges[i].Stop])))
				fmt.Printf("%s", string(line[indexRanges[i].Stop:indexRanges[i+1].Start]))
			}
			lastRangeIndex := len(indexRanges) - 1
			fmt.Printf("%s", matchedTextColor(string(line[indexRanges[lastRangeIndex].Start:indexRanges[lastRangeIndex].Stop])))
			if indexRanges[lastRangeIndex].Stop < len(line) {
				fmt.Printf("%s", string(line[indexRanges[lastRangeIndex].Stop:]))
			}
		} else if expOptions.IsInvertExpression && !isMatched { // invert expression
			printedCount += 1

			if usePrintPrefix {
				fmt.Printf("%s:", purpleTextColor(printPrefix))
			}
			fmt.Printf("%s", lineString)
		}

	}
	if printedCount > 0 && lastExtraNewLine {
		fmt.Println()
	}

	return printedCount
}

func processReader(expression string, r io.Reader, expOptions ExpressionOption, printPrefix string) (int, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024) // increase token buffer per line

	// include the newline separator
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line
			return i + 1, data[0 : i+1], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it
		if atEOF {
			return len(data), data, nil
		}
		// Request more data
		return 0, nil, nil
	})

	printedCount := 0
	switch expression {
	case `\d`, "[[:digit:]]":
		printedCount = scanExpressionPattern(scanner, containsDigit, expression, expOptions, printPrefix, true)
	case `\w`:
		printedCount = scanExpressionPattern(scanner, containsWordCharacter, expression, expOptions, printPrefix, false)
	default:
		if strings.HasPrefix(expression, "^") {
			printedCount = scanExpressionPattern(scanner, containsBeginningExpression, expression[1:], expOptions, printPrefix, false)
		} else if strings.HasSuffix(expression, "$") {
			printedCount = scanExpressionPattern(scanner, containsEndingExpression, expression[:len(expression)-1], expOptions, printPrefix, false)
		} else {
			printedCount = scanExpressionPattern(scanner, containsExpression, expression, expOptions, printPrefix, false)
		}
	}

	return printedCount, nil
}

func processOneFile(expression string, filepath string, expOptions ExpressionOption, printFilepathAsPrefix bool) (int, error) {
	fp, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	printPrefix := ""
	if printFilepathAsPrefix {
		printPrefix = filepath
	}

	return processReader(expression, fp, expOptions, printPrefix)
}

func processStream(expression string, expOptions ExpressionOption) (int, error) {
	return processReader(expression, os.Stdin, expOptions, "")
}

func main() {
	var args Args
	isOkay := args.Parse()
	if !isOkay {
		os.Exit(0)
	}

	expOptions := ExpressionOption{IsInvertExpression: args.IsInvertExpression, IsCaseInsensitive: args.IsCaseInsensitive}

	isPrinted := false

	if args.UseStdInStream {
		printedCount, err := processStream(args.Expression, expOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", args.ExeName, err)
		} else if printedCount > 0 {
			isPrinted = true
		}
	} else { // grep all files in args.Filepaths
		usePrintPrefix := false
		if len(args.Filepaths) > 1 {
			usePrintPrefix = true
		}
		for _, file := range args.Filepaths {

			if isFileExistsAndRegular(file) {
				matchedCount, err := processOneFile(args.Expression, file, expOptions, usePrintPrefix)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s: %v\n", args.ExeName, file, err)
				} else if matchedCount > 0 {
					isPrinted = true
				}
				continue
			}

			isFolder, err := isFolderExists(file)
			if !isFolder {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s: %v\n", args.ExeName, file, err)
				} else { // actually exists but not a regular folder
					fmt.Printf("%s: %s: Is not a regular file or directory\n", args.ExeName, file)
				}
				continue
			} // is a folder

			rootDir := file
			if args.IsRecurse {
				err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error on path %q: %v\n", path, err)
						return err
					}

					if isFileExistsAndRegular(path) {
						matchedCount, err := processOneFile(args.Expression, path, expOptions, usePrintPrefix)
						if err != nil {
							fmt.Fprintf(os.Stderr, "%s: %s: %v\n", args.ExeName, path, err)
						} else if matchedCount > 0 {
							isPrinted = true
						}
					}

					return nil
				})

				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
			} else {
				fmt.Printf("%s: %s: Is a directory\n", args.ExeName, file)
			}
		}
	}

	if isPrinted {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
