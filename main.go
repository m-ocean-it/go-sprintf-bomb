package main

import (
	"strconv"
	"strings"
)

func main() {
	// s := "Hello, %s!"
	// res :=
}

type SplitConcatedString struct {
	parts []string
}

func (s *SplitConcatedString) Fill(args []string) string {
	if len(args) != len(s.parts)-1 {
		panic("wrong number of args")
	}

	b := strings.Builder{}
	for i, p := range s.parts {
		if i > 0 {
			b.WriteString(" + ")
			b.WriteString(args[i-1])
			b.WriteString(" + ")
		}
		b.WriteString(p)
	}

	return b.String()
}

func SplitConcat(source string) *SplitConcatedString {
	// TODO: account for escaping

	// sep := "%s" // TODO: support more verbs
	var parts []string

	sourceRunes := []rune(source)

	var percent bool
	var nextRuneIdx int

	for i, r := range sourceRunes {
		if r == 's' && percent {
			parts = append(parts, strconv.Quote(string(sourceRunes[nextRuneIdx:i-1])))
			nextRuneIdx = i + 1
		}
		if r == '%' {
			percent = true
		} else {
			percent = false
		}
	}
	if nextRuneIdx < len(sourceRunes) {
		parts = append(parts, strconv.Quote(string(sourceRunes[nextRuneIdx:])))
	}

	return &SplitConcatedString{
		parts: parts,
	}
}
