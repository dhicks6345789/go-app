# Thin wrapper around build.sh so that `make build-all` and
# `make dist DEST_DIR=...` continue to work for existing users.
.PHONY: all build build-all index api-docs dist run-desktop run-server test clean

all: build

build:
	./build.sh build

build-all:
	./build.sh build-all

index:
	./build.sh index

api-docs:
	./build.sh api-docs

dist:
	./build.sh dist

run-desktop:
	./build.sh run-desktop

run-server:
	./build.sh run-server

test:
	./build.sh test

clean:
	./build.sh clean
