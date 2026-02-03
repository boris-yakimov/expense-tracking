compile:
	rm -f bin/expense-tracking-*
	# Native Linux amd64 build
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o bin/expense-tracking-linux-amd64 .
	# Linux ARM64 build (requires: sudo pacman -S aarch64-linux-gnu-gcc)
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/expense-tracking-linux-arm64 .
	# Windows amd64 build (requires: sudo pacman -S mingw-w64-gcc)
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/expense-tracking-windows-amd64.exe .

build:
	go build -o bin/expense-tracking .

run:
	go run .

test:
	go test -v .
