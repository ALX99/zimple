NAME="zimple"

install: build
	cp ./build/${NAME} /usr/local/bin

build:
	mkdir -p ./build
	go build -ldflags "-s -w" -o ./build/${NAME} ./cmd/zimple/zimple.go

run:
	go run ./cmd/zimple/zimple.go

clean:
	rm -rf ./build
