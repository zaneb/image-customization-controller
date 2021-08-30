ifeq (/,${HOME})
GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache/
else
GOLANGCI_LINT_CACHE=${HOME}/.cache/golangci-lint
endif
GOLANGCI_LINT ?= GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) go run github.com/golangci/golangci-lint/cmd/golangci-lint

IMG ?= image-customization-controller:latest

# Set VERBOSE to -v to make tests produce more output
VERBOSE ?= ""

all: image-customization-controller

test: generate lint unit

unit:
	go test $(VERBOSE) ./... -coverprofile cover.out

image-customization-controller: generate
	go build -o bin/image-customization-controller main.go

run:
	go run ./main.go

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: generate
generate:
	# go generate -x ./...
	$(GOLANGCI_LINT) run --fix

.PHONY: docker
docker: generate
	docker build . -t ${IMG}

.PHONY: docker-push
docker-push:
	docker push ${IMG}
