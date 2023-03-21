package module

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
)

var (
	ErrNotSupport = "not support operation"
)

const DocSiteUrl = "http://selefra.io/docs"

// RenderErrorTemplate Output Example:
//
// error[E827890]: syntax error, do not support modules[1].output
//
//	 -->  test_data\test.yaml:83:7 ( modules[1].output )
//	| 78   - name: example_module
//	| 79     uses: ./rules/
//	| 80     input:
//	| 81       name: selefra
//	| 82     output:
//	| 83       - "This is a test output message, resource region is {{.region}}."
//	|          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//	| 84
//	| 85 variables:
//	| 86   - key: test
//	| 87     default:
func RenderErrorTemplate(errorType string, location *NodeLocation) string {
	s := strings.Builder{}

	s.WriteString(fmt.Sprintf("%s: %s \n", color.RedString("error[E827890]"), errorType))
	if location == nil {
		return s.String()
	}
	s.WriteString(fmt.Sprintf("%s %s:%d:%d ( %s ) \n", color.BlueString(" --> "), location.Path, location.Begin.Line, location.Begin.Column, location.YamlSelector))

	file, err := os.ReadFile(location.Path)
	if err != nil {
		// TODO
		return err.Error()
	}
	split := strings.Split(string(file), "\n")
	// The number of characters used for lines depends on the actual number of lines in the file
	lineWidth := strconv.Itoa(len(strconv.Itoa(len(split))))
	for lineIndex, lineString := range split {
		// There can be a newline problem on Windows platforms
		lineString = strings.TrimRight(lineString, "\r")
		realLineIndex := lineIndex + 1
		// Go ahead and back a few more lines
		cutoff := 5
		if realLineIndex >= location.Begin.Line && realLineIndex <= location.End.Line {
			begin := 0
			end := len(lineString) + 1
			if realLineIndex == location.Begin.Line {
				begin = location.Begin.Column - 1
			}
			if realLineIndex == location.End.Line {
				end = location.End.Column - 1
			}

			//s.WriteString(fmt.Sprintf("| %"+lineWidth+"d: ", realLineIndex))
			s.WriteString(fmt.Sprintf("| %-"+lineWidth+"d ", realLineIndex))
			s.WriteString(lineString)
			s.WriteString("\n")

			// Error underlining
			underline := withUnderline(lineString, begin, end)
			if underline != "" {
				s.WriteString(fmt.Sprintf("| %"+lineWidth+"s ", " "))
				s.WriteString(color.RedString(underline))
				s.WriteString("\n")
			}

		} else if (realLineIndex >= location.Begin.Line-cutoff && realLineIndex < location.Begin.Line) || (realLineIndex > location.End.Line && realLineIndex <= location.End.Line+cutoff) {
			//s.WriteString(fmt.Sprintf("| %"+lineWidth+"d: ", realLineIndex))
			s.WriteString(fmt.Sprintf("| %-"+lineWidth+"d ", realLineIndex))
			s.WriteString(lineString)
			s.WriteString("\n")
		}
	}
	s.WriteString("--> See our docs: " + DocSiteUrl + "\n")

	return s.String()
}

// Underline the lines in red
func withUnderline(line string, begin, end int) string {
	underline := make([]string, 0)
	for index, _ := range line {
		if index >= begin && index <= end {
			underline = append(underline, color.RedString("^"))
		} else {
			underline = append(underline, color.RedString(" "))
		}
	}
	if len(underline) == 0 {
		return ""
	}
	return strings.Join(underline, "")
}
