REGISTRY         ?= docker.io
ORG              ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     ?= $(REGISTRY)/$(ORG)/origin-ansible-service-broker
PUSH_IMAGE       ?= 0
VARS             ?= ""
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
PREFIX           ?= /usr/local
BROKER_CONFIG    ?= $(PWD)/etc/generated_local_development.yaml
SOURCE_DIRS      = cmd pkg
SOURCES          := $(shell find . -name '*.go' -not -path "*/vendor/*")
PACKAGES         := $(shell go list ./pkg/...)
SVC_ACCT_DIR     := /var/run/secrets/kubernetes.io/serviceaccount
KUBERNETES_FILES := $(addprefix $(SVC_ACCT_DIR)/,ca.crt token tls.crt tls.key)
COVERAGE_SVC     := travis-ci
.DEFAULT_GOAL    := build

vendor: ## Install or update project dependencies
	@dep ensure

broker: $(SOURCES) ## Build the broker
	go build -i -ldflags="-s -w" ./cmd/broker

migration: $(SOURCES)
	go build -i -ldflags="-s -w" ./cmd/migration

build: broker ## Build binary from source
	@echo > /dev/null

generate: ## regenerate mocks
	go get github.com/vektra/mockery/.../
	@go generate ./...

lint: ## Run golint
	@golint -set_exit_status $(addsuffix /... , $(SOURCE_DIRS))

fmt: ## Run go fmt
	@gofmt -d $(SOURCES)

fmtcheck: ## Check go formatting
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

test: ## Run unit tests
	@go test -cover ./pkg/...

coverage-all.out: $(SOURCES)
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

test-coverage-html: coverage-all.out ## checkout the coverage locally of your tests
	@go tool cover -html=coverage-all.out

ci-test-coverage: coverage-all.out ## CI test coverage, upload to coveralls
	@goveralls -coverprofile=coverage-all.out -service $(COVERAGE_SVC)

vet: ## Run go vet
	@go tool vet ./cmd ./pkg

check: fmtcheck vet lint build test ## Pre-flight checks before creating PR

run: broker
	@./scripts/run_local.sh ${BROKER_CONFIG}

# NOTE: Must be explicitly run if you expect to be doing local development
# Basically brings down the broker pod and extracts token/cert files to
# /var/run/secrets/kubernetes.io/serviceaccount
# Resetting a catasb cluster WILL generate new certs, so you will have to
# run prep-local again to export the new certs.
prep-local: ## Prepares the local dev environment
	@./scripts/prep_local_devel_env.sh

build-image: ## Build a docker image with the broker binary, PUSH_IMAGE=1 will push it to docker
	env GOOS=linux go build -i -ldflags="-s -s" -o ${BUILD_DIR}/broker ./cmd/broker
	env GOOS=linux go build -i -ldflags="-s -s" -o ${BUILD_DIR}/migration ./cmd/migration
	docker build -f ${BUILD_DIR}/Dockerfile-localdev -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
ifneq ($(PUSH_IMAGE),0)
	docker push ${BROKER_IMAGE}:${TAG}
else
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    docker push ${BROKER_IMAGE}:${TAG}"
endif

release-image:
	docker build -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    make push"

# https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker-latest/
release: release-image ## Builds docker container using latest rpm from Copr

push:
	docker push ${BROKER_IMAGE}:${TAG}

clean: ## Clean up your working environment
	@rm -f broker
	@rm -f migration
	@rm -f build/broker
	@rm -f build/migration
	@rm -f adapters.out apb.out app.out auth.out broker.out coverage-all.out coverage.out handler.out registries.out validation.out

really-clean: clean cleanup-ci ## Really clean up the working environment
	@rm -f $(KUBERNETES_FILES)

deploy: ## Deploy a built broker docker image to a running cluster
	@./scripts/deploy.sh ${BROKER_IMAGE}:${TAG} ${REGISTRY} ${ORG} ${VARS}

## Continuous integration stuff

cleanup-ci: ## Cleanup after ci run
	./scripts/broker-ci/cleanup-ci.sh

ci: ## Run the CI workflow locally
	@go get github.com/rthallisey/service-broker-ci/cmd/ci
	@ci

ci-k:
	@go get github.com/rthallisey/service-broker-ci/cmd/ci
	@KUBERNETES="k8s" ci --cluster kubernetes

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: run build-image release-image release push clean deploy ci cleanup-ci lint build vendor fmt fmtcheck test vet help test-cover-html prep-local
