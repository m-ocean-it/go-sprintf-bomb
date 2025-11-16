package analyzer

import (
	"fmt"
	"testing"
)

func TestPrecedenceErrorOverStringer(t *testing.T) {
	t.Parallel()

	s := stringerError("internal value")

	got := fmt.Sprintf("val: %s", s)

	expected := "val: Error()"
	if got != expected {
		t.Fatalf("got: %s, expected: %s", got, expected)
	}
}

func TestPrecedenceErrorOverInternalValue(t *testing.T) {
	t.Parallel()

	s := stringError("internal value")

	got := fmt.Sprintf("val: %s", s)

	expected := "val: Error()"
	if got != expected {
		t.Fatalf("got: %s, expected: %s", got, expected)
	}
}

func TestPrecedenceStringerOverInternalValue(t *testing.T) {
	t.Parallel()

	s := stringStringer("internal value")

	got := fmt.Sprintf("val: %s", s)

	expected := "val: String()"
	if got != expected {
		t.Fatalf("got: %s, expected: %s", got, expected)
	}
}

func TestInternalStringValueIsUsed(t *testing.T) {
	t.Parallel()

	s := wrappedString("internal value")

	got := fmt.Sprintf("val: %s", s)

	expected := "val: internal value"
	if got != expected {
		t.Fatalf("got: %s, expected: %s", got, expected)
	}
}

type stringStringer string

func (s stringStringer) String() string {
	return "String()"
}

type stringerError string

func (s stringerError) String() string {
	return "String()"
}
func (s stringerError) Error() string {
	return "Error()"
}

type stringError string

func (s stringError) Error() string {
	return "Error()"
}

type wrappedString string
