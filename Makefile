run:
	DEBUG=1 go run main.go

dev:
	fswatch -e ".*" -i "\\.go$$" -o . | while read -r; do \
		echo "Change detected. Rebuilding..."; \
		go build -o build/viscue . && pkill -f 'build/viscue' || true; \
		DEBUG=1 ./build/viscue & \
	done

