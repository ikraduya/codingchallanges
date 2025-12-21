package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"ccgrep/internal/match"
	"ccgrep/internal/output"
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

func printHelp(exeName string) {
	fmt.Printf("Usage: %s [OPTION]... EXPRESSION [FILE]...\n", exeName)
	fmt.Println("OPTION:")
	fmt.Println("\t'-r' Recurse the directory tree")
	fmt.Println("\t'-v' Inverse the match expression")
	fmt.Println("\t'-i' Case insensitive match")
}

func (args *Args) Parse() (isValid bool) {
	args.ExeName = filepath.Base(os.Args[0])

	fs := flag.NewFlagSet(args.ExeName, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // prevent flag from printing parse error

	fs.BoolVar(&args.IsRecurse, "r", false, "recurse the directory tree")
	fs.BoolVar(&args.IsInvertExpression, "v", false, "invert the match expression")
	fs.BoolVar(&args.IsCaseInsensitive, "i", false, "case insensitive match")

	if err := fs.Parse(os.Args[1:]); err != nil {
		printHelp(args.ExeName)
		return false
	}

	positionals := fs.Args()
	if len(positionals) == 0 {
		fmt.Println("EXPRESSION is required")
		printHelp(args.ExeName)
		return false
	}

	args.Expression = positionals[0]
	args.Filepaths = positionals[1:]
	args.UseStdInStream = len(args.Filepaths) == 0

	return true
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

func scanExpressionPattern(scanner *bufio.Scanner, checkContainsFunc match.CheckContainsOperation, exp string, expOptions match.ExpressionOption, printPrefix string, lastExtraNewLine bool) int {
	textColorTheme := output.DefaultTextColorTheme()
	if exp == "" {
		textColorTheme.Match = output.NormalTextColor()
	}

	expRunes := []rune(exp)

	printedCount := 0
	for scanner.Scan() {
		lineString := scanner.Text()
		line := []rune(lineString)

		indexRanges := checkContainsFunc(line, expRunes, expOptions)
		isMatched := indexRanges != nil

		if !expOptions.IsInvertExpression && isMatched {
			printedCount += 1

			output.Output(line, indexRanges, printPrefix, textColorTheme)
		} else if expOptions.IsInvertExpression && !isMatched { // invert expression
			printedCount += 1

			output.OutputDefaultColor(lineString, printPrefix, textColorTheme)
		}

	}
	if printedCount > 0 && lastExtraNewLine {
		output.OutputExtraLine()
	}

	return printedCount
}

func processReader(expression string, r io.Reader, expOptions match.ExpressionOption, printPrefix string) (int, error) {
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
		printedCount = scanExpressionPattern(scanner, match.ContainsDigit, expression, expOptions, printPrefix, true)
	case `\w`:
		printedCount = scanExpressionPattern(scanner, match.ContainsWordCharacter, expression, expOptions, printPrefix, false)
	default:
		if strings.HasPrefix(expression, "^") {
			printedCount = scanExpressionPattern(scanner, match.ContainsBeginningExpression, expression[1:], expOptions, printPrefix, false)
		} else if strings.HasSuffix(expression, "$") {
			printedCount = scanExpressionPattern(scanner, match.ContainsEndingExpression, expression[:len(expression)-1], expOptions, printPrefix, false)
		} else {
			printedCount = scanExpressionPattern(scanner, match.ContainsExpression, expression, expOptions, printPrefix, false)
		}
	}

	return printedCount, nil
}

func processOneFile(expression string, filepath string, expOptions match.ExpressionOption, printFilepathAsPrefix bool) (int, error) {
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

func processStream(expression string, expOptions match.ExpressionOption) (int, error) {
	return processReader(expression, os.Stdin, expOptions, "")
}

func main() {
	var args Args
	isOkay := args.Parse()
	if !isOkay {
		os.Exit(0)
	}

	expOptions := match.ExpressionOption{IsInvertExpression: args.IsInvertExpression, IsCaseInsensitive: args.IsCaseInsensitive}

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
