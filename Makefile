REGISTRY         ?= docker.io
PROJECT          ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     = $(REGISTRY)/$(PROJECT)/ansible-service-broker
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"

install: $(shell find cmd pkg)
	go install -ldflags="-s -w" ./cmd/broker

${GOPATH}/bin/mock-registry: $(shell find cmd/mock-registry)
	go install ./cmd/mock-registry

# Will default run to dev profile
run: install vendor
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/scripts/runbroker.sh dev

deploy:
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/scripts/deploy.sh

run-mock-registry: ${GOPATH}/bin/mock-registry vendor
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/cmd/mock-registry/run.sh

prepare-build-image: install
	cp "${GOPATH}"/bin/broker build/broker

build-image: prepare-build-image
	docker build -f ${BUILD_DIR}/Dockerfile-src ${BUILD_DIR} -t ${BROKER_IMAGE}:${TAG}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    docker push ${BROKER_IMAGE}:${TAG}"

release-image:
	docker build ${BUILD_DIR} -t ${BROKER_IMAGE}:${TAG}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    make push"

push:
	docker push ${BROKER_IMAGE}:${TAG}

#Use release instead of release-image. We may add push to this later.
release: release-image

clean:
	@rm -f ${GOPATH}/bin/broker
	@rm -f build/broker

vendor:
	@glide install -v

test: vendor
	go test ./pkg/...

.PHONY: build run run-mock-registry clean test build asb-image install prepare-build-image build-image release-image push release
