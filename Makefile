REGISTRY         ?= docker.io
PROJECT          ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     ?= $(REGISTRY)/$(PROJECT)/ansible-service-broker
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
PREFIX           ?= /usr/local
BROKER_CONFIG    ?= $(PWD)/etc/generated_local_development.yaml

SOURCES := $(shell find . -name '*.go' -not -path "*/vendor/*")

vendor:
	@glide install -v

broker: $(SOURCES)
	go build -i -ldflags="-s -w" ./cmd/broker

build: broker
	@echo > /dev/null

run: broker
	cd scripts && ./run_local.sh ${BROKER_CONFIG}

prepare-local-env:
	cd scripts && ./prep_local_devel_env.sh

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

deploy:
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/scripts/deploy.sh

test:
	go test ./pkg/...

.PHONY: vendor build run prepare-build-image build-image release-image release push clean deploy test
