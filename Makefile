run:
	DEBUG=1 go run main.go

build:
	GOOS=darwin GOARCH=arm64 go build -o build/viscue .

dev:
	fswatch -e ".*" -i "\\.go$$" -o . | while read -r; do \
		echo "Change detected. Rebuilding..."; \
		go build -o build/viscue . && pkill -f 'build/viscue' || true; \
		DEBUG=1 ./build/viscue & \
	done

release:
	GITHUB_TOKEN=$(GITHUB_TOKEN) CGO_ENABLED=1 goreleaser release --clean