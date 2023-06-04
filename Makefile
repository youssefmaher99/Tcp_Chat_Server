build:
	go build -o bin/chat_server

run: build
	./bin/chat_server