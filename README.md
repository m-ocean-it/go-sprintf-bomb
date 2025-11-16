# go-sprintf-bomb

*Optimize `Sprintf`s away!*

## About

It's a linter that tries to replace `fmt.Sprintf` calls with more efficient string concatenation. (Work in progress.)

## Example

Before applied fixes:
```go
name := "John"
age := 3
pi := 3.14
err := errors.New("some error")
_ = fmt.Sprintf("%s is %d years old. Pi is %f. And some error: %s", name, age, pi, err)
```
After applied fixes:
```go
name := "John"
age := 3
pi := 3.14
err := errors.New("some error")
_ = name + " is " + strconv.Itoa(age) + " years old. Pi is " + strconv.FormatFloat(pi, 'f', -1, 64) + ". And some error: " + err.Error()
```

More examples of possible transformations can be found in the `analyzer/testdata/src/default/p.go.golden` file.

## Installation
```sh
go install github.com/m-ocean-it/go-sprintf-bomb@latest
```

## Run
```sh
git stash                    # to avoid losing changes
go-sprintf-bomb --fix ./...  # to apply all fixes
```

## Issues

- Incomplete. A lot of cases would simply be skipped.
- The longer the string and the more placeholders are used in a `Sprintf`-call, the less significant the optimization would be. It might be useful to add some heuristics to the linter to avoid actually harming the performance.

## TODO

- [ ] Add tests for comparing the resulting strings to using fmt.Sprintf. The strings must be the same.
- [x] Fix imports automatically.
