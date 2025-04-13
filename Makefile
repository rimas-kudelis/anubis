NODE_MODULES = node_modules
VERSION := $(shell cat ./VERSION)

export RUSTFLAGS=-Ctarget-feature=+simd128

.PHONY: build assets deps lint prebaked-build test wasm

assets:
	npm run assets

deps:
	npm ci
	go mod download

build: deps
	npm run build
	@echo "Anubis is now built to ./var/anubis"

all: build

lint:
	go vet ./...
	go tool staticcheck ./...

prebaked-build:
	go build -o ./var/anubis -ldflags "-X 'github.com/TecharoHQ/anubis.Version=$(VERSION)'" ./cmd/anubis

test:
	npm run test

wasm:
	cargo build --release --target wasm32-unknown-unknown
	cp -vf ./target/wasm32-unknown-unknown/release/*.wasm ./web/static/wasm