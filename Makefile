.PHONY: lint test run

# Target to run tests
test:
	go test ./sdk -v

# Target to run the demo
run:
	go run demo/main.go

lint:
	golangci-lint run