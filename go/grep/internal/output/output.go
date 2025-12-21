package output

import (
	"fmt"
	"strings"

	"ccgrep/internal/match"

	"github.com/fatih/color"
)

type TextColorTheme struct {
	Prefix func(a ...interface{}) string
	Match  func(a ...interface{}) string
	Normal func(a ...interface{}) string
}

func NormalTextColor() func(a ...interface{}) string {
	return func(a ...interface{}) string { return fmt.Sprint(a...) }
}

func DefaultTextColorTheme() TextColorTheme {
	return TextColorTheme{
		Prefix: color.New(color.FgHiMagenta).SprintFunc(),
		Match:  color.New(color.FgHiRed, color.Bold).SprintFunc(),
		Normal: NormalTextColor(),
	}
}

func OutputDefaultColor(lineString string, printPrefix string, textColorTheme TextColorTheme) {
	var sb strings.Builder

	if printPrefix != "" {
		sb.WriteString(textColorTheme.Prefix(printPrefix))
	}
	sb.WriteString(textColorTheme.Normal(lineString))

	fmt.Print(sb.String())
}

func Output(line []rune, indexRanges []match.IndexRange, printPrefix string, textColorTheme TextColorTheme) {
	var sb strings.Builder

	if printPrefix != "" {
		sb.WriteString(textColorTheme.Prefix(printPrefix))
	}

	if indexRanges[0].Start > 0 {
		sb.WriteString(textColorTheme.Normal(string(line[:indexRanges[0].Start])))
	}
	for i := 0; i < len(indexRanges)-1; i++ {
		sb.WriteString(textColorTheme.Match(string(line[indexRanges[i].Start:indexRanges[i].Stop])))
		sb.WriteString(textColorTheme.Normal(string(line[indexRanges[i].Stop:indexRanges[i+1].Start])))
	}
	lastRangeIndex := len(indexRanges) - 1
	sb.WriteString(textColorTheme.Match(string(line[indexRanges[lastRangeIndex].Start:indexRanges[lastRangeIndex].Stop])))
	if indexRanges[lastRangeIndex].Stop < len(line) {
		sb.WriteString(textColorTheme.Normal(string(line[indexRanges[lastRangeIndex].Stop:])))
	}

	fmt.Print(sb.String())
}

func OutputExtraLine() {
	fmt.Println()
}
