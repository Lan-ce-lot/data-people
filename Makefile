.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go mod tidy
	go build -o bin/data-people .

.PHONY: run
run:
	go mod tidy
	go run main.go -config test_config.yaml

.PHONY: clean
clean:
	rm -rf ./bin
	rm -rf ./data