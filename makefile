NAME := teams-help-request
OWNER := byuoitav
PKG := github.com/${OWNER}/${NAME}
DOCKER_URL := docker.pkg.github.com
DOCKER_PKG := ${DOCKER_URL}/${OWNER}/${NAME}

# version:
# use the git tag, if this commit
# doesn't have a tag, use the git hash
COMMIT_HASH := $(shell git rev-parse --short HEAD)
TAG := $(shell git rev-parse --short HEAD)
ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
	TAG = $(shell git describe --exact-match --tags HEAD)
endif

PRD_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+"
DEV_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+-.+"

# go stuff
PKG_LIST := $(shell go list ${PKG}/...)

.PHONY: all deps build test test-cov clean

all: clean build

test:
	@go test -v ${PKG_LIST}

test-cov:
	@go test -coverprofile=coverage.txt -covermode=atomic ${PKG_LIST}

lint:
	@golangci-lint run --tests=false

deps:
	@echo Downloading dependencies...
	@go mod download

build: deps
	@mkdir -p dist

	@echo
	@echo Building teams-help-request for linux-amd64
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/teams-help-request-linux-amd64

	@echo
	@echo Build output is located in ./dist/.

docker: clean build
	@echo Building prd containers with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/teams-help-request:${TAG}
	@docker build -f dockerfile --build-arg NAME=teams-help-request-linux-amd64 -t ${DOCKER_PKG}/teams-help-request:${TAG} dist


deploy: docker
	@echo Logging into Github Package Registry
	@docker login ${DOCKER_URL} -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

	@echo Pushing prd containers with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/teams-help-request:${TAG}
	@docker push ${DOCKER_PKG}/teams-help-request:${TAG}


clean:
	@go clean
	@rm -rf dist/
