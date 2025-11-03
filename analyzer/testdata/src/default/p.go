package p

import "fmt"

func foo() {
	i64 := int64(2)
	_ = fmt.Sprintf("High %d!", i64) // want "foobar"

	i := 2
	_ = fmt.Sprintf("%d is int", i) // want "foobar"

	_ = fmt.Sprintf("%s, %s, %s", "a", "b", "c") // want "foobar"

	_ = fmt.Sprintf("%s is %d years old. Pi is %f", "John", 3, 3.14) // want "foobar"

	f32 := float32(3.14)
	_ = fmt.Sprintf("Pi is %f", f32) // want "foobar"
}
