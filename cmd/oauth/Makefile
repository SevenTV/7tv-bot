.PHONY: build lint deps dev_deps generate portal clean work dev

BUILDER := "unknown"
VERSION := "unknown"

build:
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o out/oauth cmd/oauth/*.go

lint:
	golangci-lint run --go=1.18
	yarn prettier --check .

format:
	gofmt -s -w .
	yarn prettier --write .

deps:
	go mod download

dev_deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0
	yarn

generate:
	echo ${DOCROOT}

	make format

portal:
	yarn --cwd ./portal
	yarn --cwd ./portal build

portal_stage:
	yarn --cwd ./portal
	yarn --cwd ./portal build --mode=stage

test:
	#TODO: fix test
	#go test -count=1 -cover -parallel $$(nproc) -race ./...

clean:
	rm -rf \
		out \

work:
	echo -e "go 1.18\n\nuse (\n\t.\n\t../Common\n\t../message-queue/go\n\t../image-processor/go\n\t../CompactDisc\n)" > go.work
	go mod tidy

dev:
	go run cmd/oauth/main.go