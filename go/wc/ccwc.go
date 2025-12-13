package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
)

type Args struct {
	ByteCountMode      bool   `arg:"-c" help:"Byte count mode"`
	LineCountMode      bool   `arg:"-l" help:"Line count mode"`
	WordCountMode      bool   `arg:"-w" help:"Word count mode"`
	CharacterCountMode bool   `arg:"-m" help:"Character count mode"`
	Filepath           string `arg:"positional"`
}

type Result struct {
	ByteCount      int64
	LineCount      int64
	WordCount      int64
	CharacterCount int64
}

// For byte counter
type countWriter struct {
	n *int64
}

func (w countWriter) Write(p []byte) (int, error) {
	*w.n += int64(len(p))
	return len(p), nil
}

func countLineAndWord(r io.Reader) (int64, int64, error) {
	var lineCounter int64
	var wordCounter int64

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024) // increase buffer size to avoid 64K token limit per line of Scanner. Increase up to 1mil token

	for scanner.Scan() {
		line := scanner.Text()

		lineCounter += 1
		wordCounter += int64(len(strings.Fields(line)))
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	return lineCounter, wordCounter, nil
}

func countCharacter(r io.Reader) (int64, error) {
	var characterCounter int64
	reader := bufio.NewReader(r)

	for {
		_, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return characterCounter, nil
			}
			return 0, err
		}

		characterCounter += 1
	}
}

func main() {
	var args Args
	arg.MustParse(&args)

	var err error = nil
	var result Result

	isNoOptionProvided := !(args.ByteCountMode || args.LineCountMode || args.WordCountMode || args.CharacterCountMode)
	isUseStdInStream := (args.Filepath == "")

	// Counting
	if args.CharacterCountMode {
		var fp *os.File
		if isUseStdInStream {
			fp = os.Stdin
		} else {
			fp, err = os.Open(args.Filepath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			defer fp.Close()
		}

		result.CharacterCount, err = countCharacter(fp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	} else {
		isByteCountModeNecessary := (isNoOptionProvided || args.ByteCountMode)
		var reader io.Reader

		if isUseStdInStream {
			if isByteCountModeNecessary {
				reader = io.TeeReader(os.Stdin, countWriter{&result.ByteCount}) // use tee to count bytes while scanning later
			} else {
				reader = os.Stdin
			}
		} else {
			fp, err := os.Open(args.Filepath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			if isByteCountModeNecessary {
				result.ByteCount, err = io.Copy(io.Discard, fp) // count bytes first
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}

				// close and reopen (most safest)
				fp.Close()
				fp, err = os.Open(args.Filepath)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}
			}
			reader = fp

			defer fp.Close()
		}

		result.LineCount, result.WordCount, err = countLineAndWord(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}

	// Printing
	if isNoOptionProvided {
		fmt.Printf("%7d %7d %7d %s\n", result.LineCount, result.WordCount, result.ByteCount, args.Filepath)
	} else {
		if args.ByteCountMode {
			fmt.Printf("%7d %s\n", result.ByteCount, args.Filepath)
		} else if args.LineCountMode {
			fmt.Printf("%7d %s\n", result.LineCount, args.Filepath)
		} else if args.WordCountMode {
			fmt.Printf("%7d %s\n", result.WordCount, args.Filepath)
		} else if args.CharacterCountMode {
			fmt.Printf("%7d %s\n", result.CharacterCount, args.Filepath)
		}
	}
}
