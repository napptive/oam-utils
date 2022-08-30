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
all: 

.PHONY: clean
# Remove build files
clean:
	@echo "Cleaining build folder: $(BUILD_FOLDER)"
	@rm -rf $(BUILD_FOLDER)

.PHONY: test
# Test all golang files in the curdir
test:
	@echo "Executing golang tests"
	@$(GO_TEST) -v ./...

.PHONY: release
release: clean
	@mkdir -p $(BUILD_FOLDER)
	@cp README.md $(BUILD_FOLDER)
	@if [ -d "deployments" ]; then \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER) bin k8s README.md; \
	elif [ -d $(BIN_FOLDER) ]; then \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER)  bin README.md; \
	else \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER)  README.md; \
	fi
	@echo "::set-output name=release_file::$(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz"
	@echo "::set-output name=release_name::$(PROJECT_NAME)_$(VERSION).tar.gz"
