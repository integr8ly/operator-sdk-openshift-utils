SHELL = /bin/bash
PROJECT = github.com/integr8ly/operator-sdk-openshift-utils
APIS = ${PROJECT}/pkg/api
MASTER_URL =

.PHONY: setup/prepare
setup/prepare:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

.PHONY: setup/dep
setup/dep:
	@dep ensure -v

.PHONY: code/check
code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

.PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: test/unit
test/unit:
	@go test -v -race -cover ./pkg/...

.PHONY: test/integration
test/integration:
	@go test -v -race -cover ./test/integration -args -master=${MASTER_URL}

.PHONY: build/api
build/api:
	@go build ${APIS}/kubernetes
	@go build ${APIS}/schemes
	@go build ${APIS}/template

.PHONY: test/smoke
test/smoke: code/check test/unit build/api
