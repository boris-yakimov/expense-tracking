compile:
	rm -f bin/expense-tracking-*
	env GOOS=linux GOARCH=amd64 go build -o bin/expense-tracking-linux-amd64 .
	env GOOS=windows GOARCH=amd64 go build -o bin/expense-tracking-windows-amd64 .
	env GOOS=linux GOARCH=arm64 go build -o bin/expense-tracking-linux-arm64 .

build:
	go build -o bin/expense-tracking .

run:
	go run .

test:
	go test -v .
