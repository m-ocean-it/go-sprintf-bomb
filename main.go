package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/m-ocean-it/go-sprintf-bomb/analyzer"
)

func main() {
	singlechecker.Main(analyzer.New())
}
