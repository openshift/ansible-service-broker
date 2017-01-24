${GOPATH}/bin/broker: $(shell find cmd pkg)
	go install ./cmd/broker

run: ${GOPATH}/bin/broker
	@${GOPATH}/bin/broker

clean:
	@rm -f ${GOPATH}/bin/broker

.PHONY: run broker clean
