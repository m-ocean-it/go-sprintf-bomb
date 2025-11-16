package analyzer

import (
	"fmt"
	"strconv"
	"testing"
)

func BenchmarkOptimization(b *testing.B) {
	name := "John"
	age := 3
	pi := 3.14
	moreText := "hello"

	b.Run("Sprintf", func(b *testing.B) {
		for b.Loop() {
			_ = fmt.Sprintf("%s is %d years old. Pi is %f. And some error: %s", name, age, pi, moreText)
		}
	})

	b.Run("Concat", func(b *testing.B) {
		for b.Loop() {
			_ = name + " is " + strconv.Itoa(age) + " years old. Pi is " + strconv.FormatFloat(pi, 'f', -1, 64) + ". And some error: " + moreText
		}
	})
}
