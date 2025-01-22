.PHONY: all test coverage clean

all:
	go build -o . ./...

test:
	go test ./...

coverage:
	go test ./...  -coverprofile=cover.out
	go tool cover -html=cover.out

clean:
	rm -f cover.out
	rm -f cql-cli
