.PHONY: all test coverage lint clean

all:
	go build -o . ./...

test:
	go test ./...

coverage:
	go test ./...  -coverprofile=cover.out
	go tool cover -html=cover.out

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

clean:
	rm -f cover.out
	rm -f cql-cli pgcql-cli
