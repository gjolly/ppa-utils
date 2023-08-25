all:
	mkdir -p build
	go build -o build ./...

clean:
	rm -rf ./build
