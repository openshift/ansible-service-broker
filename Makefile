REGISTRY         ?= docker.io
PROJECT          ?= ansibleplaybookbundle
TAG              ?= latest
BROKER_APB_IMAGE = $(REGISTRY)/$(PROJECT)/ansible-service-broker-apb
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

prepare-build: install
	cp "${GOPATH}"/bin/broker build/broker

build: prepare-build
	docker build ${BUILD_DIR} -t ${BROKER_APB_IMAGE}:${TAG}
	@echo
	@echo "Remember you need to push your image before calling make deploy"
	@echo "    docker push ${BROKER_APB_IMAGE}:${TAG}"

clean:
	@rm -f ${GOPATH}/bin/broker
	@rm -f build/broker

vendor:
	@glide install -v

test: vendor
	go test ./pkg/...

#asb-image:
#	ansible-container build
#	ansible-container push --username $(DOCKERHUB_USER) --password $(DOCKERHUB_PASS) --push-to $(REGISTRY)/$(PROJECT) --tag $(TAG)
#	ansible-container shipit openshift --pull-from $(REGISTRY)/$(PROJECT) --tag $(TAG)
#	# fix bug in ansible-container
#	sed -i 's/-TCP//g' ./ansible/roles/ansible-service-broker-openshift/tasks/main.yml
#	docker build -t $(BROKER_APB_IMAGE) .
#	docker push $(BROKER_APB_IMAGE)

.PHONY: run run-mock-registry clean test build asb-image install prepare-build
