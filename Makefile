default: build

HOSTNAME=registry.terraform.io
NAMESPACE=govini-ai
NAME=pritunl
BINARY=terraform-provider-${NAME}
VERSION=0.2.10
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

install_local: build
	mkdir -p ~/.terraform.d/plugins/terraform.example.com/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/terraform.example.com/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	go fmt ./...

lint:
	golangci-lint run

generate:
	go generate ./...

clean:
	rm -f ${BINARY}

.PHONY: build install test testacc fmt lint generate clean
