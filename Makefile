${GOPATH}/bin/broker: $(shell find cmd pkg)
	go install ./cmd/broker

# Will default run to dev profile
run: ${GOPATH}/bin/broker
	@${GOPATH}/src/github.com/fusor/ansible-service-broker/scripts/runbroker.sh dev

clean:
	@rm -f ${GOPATH}/bin/broker

#vendor:
	#@glide install

test: vendor
	go test ./pkg/...

#.PHONY: run clena vendor test
.PHONY: run clean test
