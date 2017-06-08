REGISTRY         ?= docker.io
PROJECT          ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_IMAGE     = $(REGISTRY)/$(PROJECT)/ansible-service-broker
BUILD_DIR        = "${GOPATH}/src/github.com/openshift/ansible-service-broker/build"
INSTALL_DIR      ?= /usr/local

vendor:
	@glide install -v

build: vendor
	go build -ldflags="-s -w" ./cmd/broker

install: build
	cp broker ${INSTALL_DIR}/bin/ansible-service-broker
	mkdir -p ${INSTALL_DIR}/etc/ansible-service-broker
	cp etc/ex.dev.config.yaml ${INSTALL_DIR}/etc/ansible-service-broker/config.yaml

run: install 
	${INSTALL_DIR}/bin/ansible-service-broker --config ${INSTALL_DIR}/etc/ansible-service-broker/config.yaml

uninstall:
	rm  -f ${INSTALL_DIR}/bin/ansible-service-broker
	rm -rf ${INSTALL_DIR}/etc/ansible-service-broker

build-mock-registry:
	go build -ldflags="-s -w" ./cmd/mock-registry

install-mock-registry:
	cp mock-registry ${INSTALL_DIR}/bin/mock-registry
	mkdir -p ${INSTALL_DIR}/etc/mock-registry
	cp cmd/mock-registry/playbookbundles.yaml ${INSTALL_DIR}/etc/mock-registry/playbookbundles.yaml

run-mock-registry:
	${INSTALL_DIR}/bin/mock-registry --appfile ${INSTALL_DIR}/etc/mock-registry/playbookbundles.yaml

uninstall-mock-registry:
	rm  -f ${INSTALL_DIR}/bin/mock-registry
	rm -rf ${INSTALL_DIR}/etc/mock-registry


prepare-build-image: build
	cp broker build/broker

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

release: release-image

push:
	docker push ${BROKER_IMAGE}:${TAG}
clean:
	@rm -f broker
	@rm -f build/broker
	@rm -f mock-registry 

deploy:
	@${GOPATH}/src/github.com/openshift/ansible-service-broker/scripts/deploy.sh

test: vendor
	go test ./pkg/...

.PHONY: vendor build install run uninstall build-mock-registry install-mock-registry run-mock-registry uninstall-mock-registry prepare-build-image build-image release-image release push clean deploy test
