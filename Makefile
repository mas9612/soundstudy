GOBIN := go
BINNAME := soundstudy

all: test build

.PHONY: build
build:
	$(GOBIN) build -o $(BINNAME) .

.PHONY: test
test:
	$(GOBIN) test -v ./...
