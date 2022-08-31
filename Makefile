include Makefile.const

#Config variables

#Override this varible if you want to work with a specific target.
BUILD_TARGETS?=$(CURRENT_BIN_TARGETS)

#Version must be overrided in the CI 
VERSION?=local

# Docker options
TARGET_DOCKER_REGISTRY ?= $$USER

# Kubernetes options
TARGET_K8S_NAMESPACE ?= napptive

# Variables
BUILD_FOLDER=$(CURDIR)/build

# Obtain the last commit hash
COMMIT=$(shell git log -1 --pretty=format:"%H")

# Tools
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X=main.Commit=$(COMMIT)"

UNAME := $(shell uname)

ifeq ($(UNAME), Darwin)
	SED := gsed
else
	SED := sed
endif


.PHONY: all
all: test

.PHONY: test
# Test all golang files in the curdir
test:
	@echo "Executing golang tests"
	@$(GO_TEST) -v ./...

.PHONY: coverage
# Create a coverage report for all golang files in the curdir
coverage:
	@echo "Creating golang test coverage report: $(BUILD_FOLDER)/coverage.out"
	@mkdir -p $(BUILD_FOLDER)
	@$(GO_TEST) -v ./... -coverprofile=$(BUILD_FOLDER)/cover.out