build:
	rm -rf ./bin | true
	mkdir ./bin
	go build -o ./bin/monswipe .
