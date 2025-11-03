.PHONY: build install test clean fmt lint docs

HOSTNAME=terraform-providers
NAMESPACE=truenas
NAME=truenas
BINARY=terraform-provider-${NAME}
VERSION=0.2.14
OS_ARCH=linux_amd64

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -v ./...

clean:
	rm -f ${BINARY}
	rm -rf dist/

fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

lint:
	golangci-lint run

docs:
	go generate

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

