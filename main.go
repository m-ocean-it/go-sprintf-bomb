package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"app/analyzer"
)

func main() {
	singlechecker.Main(analyzer.New())
}
