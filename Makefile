build:
	rm -rf ./bin | true
	mkdir ./bin
	go build -o ./bin/monswipe .

clean:
	rm -rf ./bin | true
