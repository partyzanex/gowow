.PHONY: gen
gen:
	go generate ./...

.PHONY: build-server
build-server: bin
	@mkdir -p ./bin && \
	go build -o ./bin/gowow-server ./cmd/gowow-server

.PHONY: build-client
build-client: bin
	@mkdir -p ./bin && \
	go build -o ./bin/gowow-client ./cmd/gowow-client

.PHONY: docker-build-server
docker-build-server:
	docker build -t gowow-server:latest -f ./cmd/gowow-server/Dockerfile .

.PHONY: docker-build-client
docker-build-client:
	docker build -t gowow-client:latest -f ./cmd/gowow-client/Dockerfile .

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0 run

.PHONY: test
test:
	go test -v -count=1 -race ./... -coverprofile=coverage.txt -covermode=atomic

.PHONY: docker-server
docker-server:
	docker run --rm -p 7700:7700 -e  gowow-server:latest

.PHONY: docker-client
docker-client:
	docker run --rm gowow-client:latest gowow-client --address=host.docker.internal:7700