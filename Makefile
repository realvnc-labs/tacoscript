# set default shell
build:
	mkdir -p bin
	go build -o bin/tacoscript main.go

test:
	go test -race -v -p 1 ./...

fmt:
	goimports -w .
	gofmt -w .

lint:
	golangci-lint run
