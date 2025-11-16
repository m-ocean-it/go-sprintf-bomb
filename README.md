# go-sprintf-bomb

*Optimize `Sprintf`s away!*

## About

It's a linter that tries to replace `fmt.Sprintf` calls with more efficient string concatenation. (Work in progress.)

## Example

Before applied fixes:
```go
import (
    "errors"
    "fmt"
)

func main() {
    name := "John"
    age := 3
    pi := wrappedFloat64(3.14)
    err := errors.New("some error")

    _ = fmt.Sprintf("%s is %d years old. Pi is %f. And some error: %s", name, age, pi, err)
}

type wrappedFloat64 float64
```
After applied fixes:
```go
import (
    "errors"
    // updated import:
    "strconv"
)

func main() {
    name := "John"
    age := 3
    pi := wrappedFloat64(3.14)
    err := errors.New("some error")

    // replaced fmt.Sprintf with string-concatenation and required transformations:
    _ = name + " is " + strconv.Itoa(age) + " years old. Pi is " + strconv.FormatFloat(/*added cast:*/ float64(pi), 'f', -1, 64) + ". And some error: " + err.Error()
}

type wrappedFloat64 float64
```

More examples of possible transformations can be found in the `analyzer/testdata/src/default/p.go.golden` file.

## Installation
```sh
go install github.com/m-ocean-it/go-sprintf-bomb@latest
```

## Run
Dry-run:
```sh
go-sprintf-bomb ./...
```

Apply all fixes:
```sh
go-sprintf-bomb --fix ./...
```

**Be careful! Applying the fixes would mofidy your files. Be sure to commit/copy them before doing so.**

(You've been warned. After all, it's a bomb...)

# Features
- Updates imports as needed.
- Does not care about the shape and size of the format-string. The formatting placeholders (`%s`, `%d`, `%f`, etc.) can be at any position in the string and in any amount.
- Knows about the `errors.error` and `fmt.Stringer` interfaces and considers them when processing the `%s` formatting directive. Assumes the precedence of those interfaces in the same way the `fmt` library does.


## Issues

- The longer the string and the more placeholders are used in a `Sprintf`-call, the less significant the optimization would be. It might be useful to add some heuristics to the linter to avoid actually harming the performance.

## TODO

- [ ] Add tests for comparing the resulting strings to using `fmt.Sprintf`. The strings must be the same.
- [ ] Support complex float-formatting (i.e. consider more directives than just the plain `%f`).