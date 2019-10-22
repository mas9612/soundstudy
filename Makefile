GOBIN := go
BINNAME := soundstudy

all: test build plot

.PHONY: build
build:
	$(GOBIN) build -o $(BINNAME) .

.PHONY: test
test:
	$(GOBIN) test -v ./...

.PHONY: plot
plot:
	./plot_sample.sh
