REGISTRY         ?= docker.io
PROJECT          ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     ?= $(REGISTRY)/$(PROJECT)/ansible-service-broker
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
PREFIX           ?= /usr/local
BROKER_CONFIG    ?= $(PWD)/etc/generated_local_development.yaml

vendor:
	@glide install -v

build:
	go build -ldflags="-s -w" ./cmd/broker


install:
	cp broker ${PREFIX}/bin/ansible-service-broker
	mkdir -p ${PREFIX}/etc/ansible-service-broker
	cp etc/example-config.yaml ${PREFIX}/etc/ansible-service-broker/config.yaml

run:
	cd scripts && ./run_local.sh ${BROKER_CONFIG}

uninstall:
	rm  -f ${PREFIX}/bin/ansible-service-broker
	rm -rf ${PREFIX}/etc/ansible-service-broker

build-mock-registry:
	go build -ldflags="-s -w" ./cmd/mock-registry

install-mock-registry:
	cp mock-registry ${PREFIX}/bin/mock-registry
	mkdir -p ${PREFIX}/etc/mock-registry
	cp cmd/mock-registry/playbookbundles.yaml ${PREFIX}/etc/mock-registry/playbookbundles.yaml

run-mock-registry:
	${PREFIX}/bin/mock-registry --appfile ${PREFIX}/etc/mock-registry/playbookbundles.yaml

uninstall-mock-registry:
	rm  -f ${PREFIX}/bin/mock-registry
	rm -rf ${PREFIX}/etc/mock-registry

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
	@rm -f mock-registry

deploy:
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/scripts/deploy.sh

test:
	go test ./pkg/...

.PHONY: vendor build install run uninstall build-mock-registry install-mock-registry run-mock-registry uninstall-mock-registry prepare-build-image build-image release-image release push clean deploy test
