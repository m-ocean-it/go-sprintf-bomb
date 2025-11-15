package p

import "fmt"

func foo() {
	i64 := int64(2)
	_ = fmt.Sprintf("High %d!", i64) // want "Sprintf could be optimized away"

	i := 2
	_ = fmt.Sprintf("%d is int", i) // want "Sprintf could be optimized away"

	_ = fmt.Sprintf("%s, %s, %s", "a", "b", "c") // want "Sprintf could be optimized away"

	_ = fmt.Sprintf("%s is %d years old. Pi is %f", "John", 3, 3.14) // want "Sprintf could be optimized away"

	f32 := float32(3.14)
	_ = fmt.Sprintf("Pi is %f", f32) // want "Sprintf could be optimized away"

	cs := customStringer{}
	_ = fmt.Sprintf("This is %s", cs) // want "Sprintf could be optimized away"
}

type customStringer struct{}

func (c customStringer) String() string {
	return "Hello from custom stringer!"
}
