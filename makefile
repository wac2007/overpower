output = build/op
build:
	go build -o $(output) .

run:
	go run .

install: build
	sudo cp $(output) /usr/bin/op