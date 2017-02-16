#${GOPATH}/bin/broker: $(shell find cmd pkg)
build: $(shell find cmd pkg)
	go install ./cmd/broker

${GOPATH}/bin/mock-registry: $(shell find cmd/mock-registry)
	go install ./cmd/mock-registry

# Will default run to dev profile
run: build vendor
	@${GOPATH}/src/github.com/fusor/ansible-service-broker/scripts/runbroker.sh dev

run-mock-registry: ${GOPATH}/bin/mock-registry vendor
	@${GOPATH}/src/github.com/fusor/ansible-service-broker/cmd/mock-registry/run.sh

clean:
	@rm -f ${GOPATH}/bin/broker

vendor:
	@glide install

test: vendor
	go test ./pkg/...

.PHONY: run run-mock-registry clean test build
