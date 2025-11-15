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
_ = fmt.Sprintf("%s is %d years old. Pi is %f", name, age, pi)
```
After applied fixes:
```go
name := "John"
age := 3
pi := 3.14
_ = name + " is " + strconv.Itoa(age) + " years old. Pi is " + strconv.FormatFloat(pi, 'f', -1, 64)
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

- Package imports aren't handled. You might need to fix them yourself after applying the fixes. (For example, a fix might introduce a dependency on `strconv`, but the package won't be automatically added to the `imports` section.)
- Incomplete. A lot of cases would be simply skipped.