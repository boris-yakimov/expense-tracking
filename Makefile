compile:
	rm -f bin/expense-tracking-*
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o bin/expense-tracking-linux-amd64 .
	# requires ARM cross compiler to be installed - sudo apt install gcc-aarch64-linux-gnu 
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/expense-tracking-linux-arm64 .
	# requires MinGW cross compiler to be installed - sudo apt install gcc-mingw-w64
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/expense-tracking-windows-amd64.exe .

build:
	go build -o bin/expense-tracking .

run:
	go run .

test:
	go test -v .
