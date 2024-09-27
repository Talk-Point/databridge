setup:
	echo "Setting up your application"

test:
	go test ./... -v

test-cover:
	go test -cover ./...

build:
	go build -o bin/databridge ./cmd/databridge/main.go

build-docker:
	docker build -t talkpoint/databridge:latest .
