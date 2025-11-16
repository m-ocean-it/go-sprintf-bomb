.PHONY: build
build:
	mkdir -p bin && go build -o bin/go-sprintf-bomb .