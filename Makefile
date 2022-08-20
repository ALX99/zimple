NAME="zimple"

install: build
	cp ./build/${NAME} /usr/local/bin

build:
	mkdir -p ./build
	go build -ldflags "-s -w" -o ./build/${NAME} ./cmd/main.go

run:
	go run ./cmd/main.go

clean:
	rm -rf ./build
