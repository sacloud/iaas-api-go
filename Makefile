AUTHOR          ?="The sacloud/iaas-api-go Authors"
COPYRIGHT_YEAR  ?="2022"
COPYRIGHT_FILES ?=$$(find . -name "*.go" -print | grep -v "/vendor/")

default: gen fmt set-license go-licenses-check goimports lint test

.PHONY: test
test:
	TESTACC= go test ./... $(TESTARGS) -v -timeout=120m -parallel=8 -race;

.PHONY: testacc
testacc:
	TESTACC=1 go test ./... $(TESTARGS) --tags=acctest -v -timeout=120m -parallel=8 ;

.PHONY: tools
tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go install github.com/sacloud/addlicense@latest
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/google/go-licenses@v1.0.0
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.44.2/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.44.2

.PHONY: clean
clean:
	find . -type f -name "zz_*.go" -delete

.PHONY: gen
gen: _gen fmt goimports set-license

.PHONY: _gen
_gen:
	go generate ./...

.PHONY: goimports
goimports: fmt
	goimports -l -w .

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

.PHONY: godoc
godoc:
	echo "URL: http://localhost:6060/pkg/github.com/sacloud/iaas-api-go/"
	godoc -http=localhost:6060

.PHONY: lint-all
lint-all: lint-go lint-text

.PHONY: lint lint-go
lint: lint-go
lint-go:
	golangci-lint run ./...

.PHONY: textlint lint-text
textlint: lint-text
lint-text:
	@docker run -it --rm -v $$PWD:/work -w /work ghcr.io/sacloud/textlint-action:v0.0.1 .

.PHONY: set-license
set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)

.PHONY: go-licenses-check
go-licenses-check:
	go-licenses check .