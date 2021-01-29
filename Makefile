PREFIX:=/usr/local/bin

all: build

build:
	mkdir -p bin
	go build -o bin/optimistic-title ./...

install:
	install -m 0755 bin/optimistic-title $(PREFIX)/optimistic-title
