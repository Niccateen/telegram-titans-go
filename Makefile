.PHONY: build clean run deps check-deps

# Binary name
BINARY=tdl-gui

# Build flags
LDFLAGS=-ldflags="-s -w"

build: deps
	go build $(LDFLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

deps:
	go mod tidy

clean:
	rm -f $(BINARY)
	go clean

# Check if required system dependencies are available
check-deps:
	@echo "Checking for required dependencies..."
	@which tdl > /dev/null 2>&1 && echo "✓ tdl found" || echo "✗ tdl not found - install from https://docs.iyear.me/tdl/"
	@which go > /dev/null 2>&1 && echo "✓ go found" || echo "✗ go not found"
	@pkg-config --exists gl 2>/dev/null && echo "✓ OpenGL found" || echo "✗ OpenGL dev libs missing (install libgl1-mesa-dev)"
	@pkg-config --exists x11 2>/dev/null && echo "✓ X11 found" || echo "✗ X11 dev libs missing (install xorg-dev)"
	@echo "Done."

# Build for different platforms (cross-compilation)
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux .

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build $(LDFLAGS) -o $(BINARY).exe .

help:
	@echo "TDL GUI Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build       - Build the GUI application"
	@echo "  run         - Build and run the application"
	@echo "  deps        - Download Go dependencies"
	@echo "  clean       - Remove built binary"
	@echo "  check-deps  - Check if system dependencies are installed"
	@echo "  help        - Show this help message"

