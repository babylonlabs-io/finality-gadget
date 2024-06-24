# Makefile

# Variables
RUSTFLAGS := -C link-arg=-s

# Default target
all: build

# Build target
build:
	cargo schema
	RUSTFLAGS='$(RUSTFLAGS)' cargo wasm

# Generate wasm
wasm:
	cargo build --release --target wasm32-unknown-unknown --lib

# Clean target
clean:
	cargo clean

.PHONY: all build clean
