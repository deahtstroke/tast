.PHONY: test
test:
	go test -v ./...

.PHONY: coverage
coverage: test
	go tool cover -html=coverage.out
	go test -v -coverprofile=coverage.out ./...
