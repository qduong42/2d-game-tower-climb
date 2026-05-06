.PHONY: dev build test clean

CLIENT_DIST=client/dist

dev:
	@echo "Starting server + client dev..."
	@cd client && npm run dev &
	@go run ./cmd/server

build: $(CLIENT_DIST)
	go build -o server ./cmd/server

$(CLIENT_DIST):
	cd client && npm ci && npm run build

test:
	go test ./...
	cd client && npm test

clean:
	rm -f server
	rm -rf client/dist
