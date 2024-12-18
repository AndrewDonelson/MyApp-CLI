# Detect operating system and shell environment
ifeq ($(OS),Windows_NT)
    ifeq ($(MSYSTEM),MINGW64)
        BINARY_NAME := myapp-cli.exe
        RM := rm -f
        DATE_CMD := date +'%Y.%m.%d.%H%M%S'
        MKDIR := mkdir -p
    else
        BINARY_NAME := myapp-cli.exe
        RM := del /Q
        DATE_CMD := powershell -Command "Get-Date -Format 'yyyy.MM.dd.HHmmss'"
        MKDIR := mkdir
    endif
else
    BINARY_NAME := myapp-cli
    RM := rm -f
    DATE_CMD := date +%Y.%m.%d.%H%M%S
    MKDIR := mkdir -p
endif

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get

# Main package location
MAIN_PACKAGE := .

# Build directory
BUILD_DIR := build

# Get version from current date/time
VERSION := $(shell $(DATE_CMD))

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test run help version

all: clean build

build: 
	@echo "Building for $(OS) in $(MSYSTEM)..."
	@$(MKDIR) $(BUILD_DIR) 2>/dev/null || exit 0
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Version: $(VERSION)"

clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@$(RM) $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

run: build
	@echo "Running application..."
	@$(BUILD_DIR)/$(BINARY_NAME)

version:
	@echo $(VERSION)

help:
	@echo "Available commands:"
	@echo "  make build   - Build the application"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make run     - Build and run the application"
	@echo "  make all     - Clean and build"
	@echo "  make version - Show current version string"
	@echo "  make help    - Show this help message"