REGISTRY         ?= docker.io
ORG              ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     ?= $(REGISTRY)/$(ORG)/origin-ansible-service-broker
VARS             ?= ""
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
PREFIX           ?= /usr/local
BROKER_CONFIG    ?= $(PWD)/etc/generated_local_development.yaml
SOURCE_DIRS      = cmd pkg
SOURCES          := $(shell find . -name '*.go' -not -path "*/vendor/*")
#PACKAGES         := $(shell find ./pkg/ -type d -not -path '*/\.*')
PACKAGES         := $(shell go list ./pkg/...)
SVC_ACCT_DIR     := /var/run/secrets/kubernetes.io/serviceaccount
KUBERNETES_FILES := $(addprefix $(SVC_ACCT_DIR)/,ca.crt token tls.crt tls.key)
.DEFAULT_GOAL    := build

vendor: ## Install or update project dependencies
	@glide install -v

broker: $(SOURCES) ## Build the broker
	go build -i -ldflags="-s -w" ./cmd/broker

build: broker ## Build binary from source
	@echo > /dev/null

lint: ## Run golint
	@golint -set_exit_status $(addsuffix /... , $(SOURCE_DIRS))

fmt: ## Run go fmt
	@gofmt -d $(SOURCES)

fmtcheck: ## Check go formatting
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

test: ## Run unit tests
	@go test -cover ./pkg/...

test-cover-html:
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	@go tool cover -html=coverage-all.out -o coverage.html

vet: ## Run go vet
	@go tool vet ./cmd ./pkg

check: fmtcheck vet lint build test ## Pre-flight checks before creating PR

run: broker | $(KUBERNETES_FILES) ## Run the broker locally, configure via etc/generated_local_development.yaml
	@./scripts/run_local.sh ${BROKER_CONFIG}

$(KUBERNETES_FILES):
	@./scripts/prep_local_devel_env.sh

prepare-local-env: $(KUBERNETES_FILES) ## Prepare the local environment for running the broker locally
	@echo > /dev/null

build-image: ## Build a docker image with the broker binary
	env GOOS=linux go build -i -ldflags="-s -s" -o ${BUILD_DIR}/broker ./cmd/broker
	docker build -f ${BUILD_DIR}/Dockerfile-localdev -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    docker push ${BROKER_IMAGE}:${TAG}"

release-image:
	docker build -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    make push"

# https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker/
release: release-image ## Builds docker container using latest rpm from Copr

push:
	docker push ${BROKER_IMAGE}:${TAG}

clean: ## Clean up your working environment
	@rm -f broker
	@rm -f build/broker

really-clean: clean cleanup-ci ## Really clean up the working environment
	@rm -f $(KUBERNETES_FILES)

deploy: ## Deploy a built broker docker image to a running cluster
	@./scripts/deploy.sh ${BROKER_IMAGE}:${TAG} ${REGISTRY} ${ORG} ${VARS}

## Continuous integration stuff

cleanup-ci: ## Cleanup after ci run
	./scripts/broker-ci/cleanup-ci.sh

ci: ## Run the CI workflow locally
	./scripts/broker-ci/local-ci.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run build-image release-image release push clean deploy ci cleanup-ci lint build vendor fmt fmtcheck test vet help test-cover-html
