.PHONY: test http-test bench lint

test:
	go test -v -coverprofile=coverage.out

http-test:
	go tool cover -html=coverage.out

bench:
	go test -bench=. -benchmem

lint:
	golangci-lint run --config .golangci.yml ./...
