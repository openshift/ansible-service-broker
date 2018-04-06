SOURCE_DIRS      = apb clients crd registries runtime
SOURCES          := $(shell find . -name '*.go' -not -path "*/vendor/*" -not -path "*/.git/*")
PACKAGES         := $(shell go list $(addprefix ./, $(addsuffix /... , $(shell ls -d */ | grep -v vendor))))
COVERAGE_SVC     := travis-ci
.DEFAULT_GOAL    := build

ensure: ## Install or update project dependencies
	@dep ensure

build: $(SOURCES) ## Build Test
	go build -i -ldflags="-s -w" ./...

lint: ## Run golint
	@$(foreach dir,$(SOURCE_DIRS),\
		golint -set_exit_status $(dir)/...;)

fmt: ## Run go fmt
	@gofmt -d $(SOURCES)

fmtcheck: ## Check go formatting
	@gofmt -l $(SOURCES) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

test: ## Run unit tests
	@go test -cover $(addprefix ./, $(addsuffix /... , $(SOURCE_DIRS)))

vet: ## Run go vet
	@$(foreach dir,$(SOURCE_DIRS),\
		go tool vet $(dir);)

coverage-all.out: $(SOURCES)
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

test-coverage-html: coverage-all.out ## Check out the test coverage locally
	@go tool cover -html=coverage-all.out

ci-test-coverage: coverage-all.out ## CI test coverage, upload to coveralls
	@goveralls -coverprofile=coverage-all.out -service $(COVERAGE_SVC)

check: fmtcheck vet lint build test ## Pre-flight checks before creating PR

clean: ## Clean up your working environment
	@rm -f coverage-all.out coverage.out

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: ensure build lint fmt fmtcheck test vet check help test-coverage-html clean
