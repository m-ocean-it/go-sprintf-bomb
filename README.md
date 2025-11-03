# go-sprintf-bomb

*Optimize `Sprintf`s away!*

## About

It's a linter that tries to replace `fmt.Sprintf` calls with more efficient string concatenation. (Work in progress.)

## Example

Before applied fixes:
```go
_ = fmt.Sprintf("%s is %d years old. Pi is %f", "John", 3, 3.14)
```
After applied fixes:
```go
_ = "John" + " is " + strconv.Itoa(3) + " years old. Pi is " + strconv.FormatFloat(3.14, 'f', -1, 64)
```
(Not optimal at the moment, since, for example, `strconv.Itoa(3)` could simply be written as part of the string, i.e. `John is 3 years old`.)

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