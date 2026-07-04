run-tests:
	go test ./... -race -count=1

test:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

bench:
	go test ./... -bench=. -benchmem -run=^$$

build:
	go build -o bin/gosched ./cmd/scheduler

serve: build
	./bin/gosched serve --http :8080 --grpc :50051

proto:
	protoc --go_out=. --go_opt=module=github.com/krwg/gosched \
		--go-grpc_out=. --go-grpc_opt=module=github.com/krwg/gosched \
		api/proto/gosched/v1/scheduler.proto

run: build
	./bin/gosched schedule --algorithm=edf --tasks=tests/fixtures/tasks.json

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else echo "golangci-lint not installed, skipping"; fi

cover-html: test
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin coverage.out coverage.html

.PHONY: build test run bench lint cover-html clean run-tests serve proto
