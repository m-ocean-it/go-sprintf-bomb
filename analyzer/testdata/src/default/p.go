package p

import "fmt"

func foo() {
	i64 := int64(2)
	_ = fmt.Sprintf("High %d!", i64) // want "foobar"

	i := 2
	_ = fmt.Sprintf("%d is int", i)

	_ = fmt.Sprintf("%s, %s, %s", "a", "b", "c") // want "foobar"
}
