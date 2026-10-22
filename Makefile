.PHONY: run test build clean swagger

run:
	go run cmd/app/main.go

test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -o bin/server cmd/app/main.go

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs

clean:
	rm -rf bin/
