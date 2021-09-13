all: build

build: 
	go build -o ./bin/experiments experiments/*.go
clean: 
	rm -rf ./bin

install:
	go install
test: 
	go test ./anns
	go test ./vec
	go test ./hash
	
refresh_github:
	rm -rf ../github.com
	go get ./cmd/server
	go get ./cmd/client
	go get ./experiments
