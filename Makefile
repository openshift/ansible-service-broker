REGISTRY         ?= docker.io
ORG              ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     ?= $(REGISTRY)/$(ORG)/ansible-service-broker
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
PREFIX           ?= /usr/local
BROKER_CONFIG    ?= $(PWD)/etc/generated_local_development.yaml
SOURCES          := $(shell find . -name '*.go' -not -path "*/vendor/*")
SVC_ACCT_DIR     := /var/run/secrets/kubernetes.io/serviceaccount
KUBERNETES_FILES := $(addprefix $(SVC_ACCT_DIR)/,ca.crt token tls.crt tls.key)
.DEFAULT_GOAL    := build

vendor:
	@glide install -v

broker: $(SOURCES)
	go build -i -ldflags="-s -w" ./cmd/broker

build: broker
	@echo > /dev/null

run: broker | $(KUBERNETES_FILES)
	@./scripts/run_local.sh ${BROKER_CONFIG}

$(KUBERNETES_FILES):
	@./scripts/prep_local_devel_env.sh

prepare-local-env: $(KUBERNETES_FILES)
	@echo > /dev/null

prepare-build-image: build
	cp broker build/broker

build-image: prepare-build-image
	docker build -f ${BUILD_DIR}/Dockerfile-src -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    docker push ${BROKER_IMAGE}:${TAG}"

release-image:
	docker build -t ${BROKER_IMAGE}:${TAG} ${BUILD_DIR}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    make push"

release: release-image

push:
	docker push ${BROKER_IMAGE}:${TAG}
clean:
	@rm -f broker
	@rm -f build/broker

really-clean: clean
	@rm -f $(KUBERNETES_FILES)

deploy:
	@./scripts/deploy.sh ${BROKER_IMAGE}:${TAG} ${REGISTRY} ${ORG}

test:
	go test ./pkg/...

.PHONY: vendor build run prepare-build-image build-image release-image release push clean deploy test
