build: $(shell find cmd pkg)
	# HACK: Unless docker's vendor directory is removed, we end up with a
	# duplicate symbol error from the linker that prevents compilation.
	rm -rf ${GOPATH}/src/github.com/fusor/ansible-service-broker/vendor/github.com/docker/docker/vendor && \
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
