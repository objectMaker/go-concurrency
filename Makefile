BINARY_NAME=concurrency-app
DSN="host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable"
REDIS="127.0.0.1:6379"

build:
	@echo "building"
	@env CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/${BINARY_NAME} ./cmd/web
	@echo "Built"

run:build
	@echo "starting"
	@env DSN=${DSN} REDIS=${REDIS} ./bin/${BINARY_NAME}

clean:
	@go clean
	@rm -rf bin
	@echo "cleaned"

stop:
	@-pkill -f "bin/${BINARY_NAME}"

start:run

restart:stop run

test:
	go test -v ./...