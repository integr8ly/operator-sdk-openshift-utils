SHELL = /bin/bash
PROJECT = github.com/integr8ly/operator-sdk-openshift-utils
APIS = ${PROJECT}/pkg/api

setup/prepare:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

setup/dep:
	@dep ensure -v

code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

test/unit:
	@go test -v -race -cover ./pkg/...

build:
	@go build ${APIS}/kubernetes
	@go build ${APIS}/schemes
	@go build ${APIS}/template

test/smoke: code/check test/unit build
