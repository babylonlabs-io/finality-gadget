# Makefile

# Variables
RUSTFLAGS := -C link-arg=-s

# Default target
all: build

# Build target
build:
	cargo schema
	RUSTFLAGS='$(RUSTFLAGS)' cargo wasm

# Clean target
clean:
	cargo clean

.PHONY: all build clean
