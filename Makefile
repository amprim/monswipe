build:
	rm ./bin | true
	mkdir ./bin
	go build -o ./bin/monswipe .