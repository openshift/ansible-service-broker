${GOPATH}/bin/broker: $(shell find cmd pkg)
	go install ./cmd/broker

run: ${GOPATH}/bin/broker
	@${GOPATH}/bin/broker

.PHONY: run broker
