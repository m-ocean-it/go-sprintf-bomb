package p

import (
	"errors"
	"fmt"
)

func foo() {
	i64 := int64(2)
	_ = fmt.Sprintf("High %d!", i64) // want "Sprintf could be optimized away"

	i32 := int32(2)
	_ = fmt.Sprintf("High %d!", i32) // want "Sprintf could be optimized away"

	i16 := int16(2)
	_ = fmt.Sprintf("High %d!", i16) // want "Sprintf could be optimized away"

	i8 := int8(2)
	_ = fmt.Sprintf("High %d!", i8) // want "Sprintf could be optimized away"

	i := 2
	_ = fmt.Sprintf("%d is int", i) // want "Sprintf could be optimized away"
	_ = fmt.Sprintf("%d", i)        // want "Sprintf could be optimized away"

	_ = fmt.Sprintf("%s, %s, %s", "a", "b", "c") // want "Sprintf could be optimized away"

	_ = fmt.Sprintf("%s is %d years old. Pi is %f", "John", 3, 3.14) // want "Sprintf could be optimized away"

	f32 := float32(3.14)
	_ = fmt.Sprintf("Pi is %f", f32) // want "Sprintf could be optimized away"

	cs := customStringer{}
	_ = fmt.Sprintf("This is %s", cs) // want "Sprintf could be optimized away"

	err := errors.New("some error")
	_ = fmt.Sprintf("this is an error: %s", err) // want "Sprintf could be optimized away"

	u := uint(10)
	_ = fmt.Sprintf(":%d", u) // want "Sprintf could be optimized away"

	u64 := uint64(5000)
	_ = fmt.Sprintf(":%d", u64) // want "Sprintf could be optimized away"

	u32 := uint32(190)
	_ = fmt.Sprintf(":%d", u32) // want "Sprintf could be optimized away"

	u16 := uint16(5)
	_ = fmt.Sprintf(":%d", u16) // want "Sprintf could be optimized away"

	u8 := uint8(150)
	_ = fmt.Sprintf(":%d", u8) // want "Sprintf could be optimized away"
}

type customStringer struct{}

func (c customStringer) String() string {
	return "Hello from custom stringer!"
}
