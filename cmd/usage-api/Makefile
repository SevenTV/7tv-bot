.PHONY: build lint deps generate

BUILDER := "unknown"
VERSION := "unknown"

build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o out/usage-api cmd/usage-api/*.go

lint:
	golangci-lint run --go=1.18

format:
	gofmt -s -w .

deps:
	go mod download

generate:
	echo ${DOCROOT}

	make format

test:
	#TODO: fix test
	#go test -count=1 -cover -parallel $$(nproc) -race ./...

clean:
	rm -rf \
		out \