export GOPRIVATE=github.com/ecumenos-social
export SHELL=/bin/sh

.PHONY: all
all: hooks submodules tidy check fmt lint test tidy

.PHONY: hooks
hooks: ## Git hooks
	git config core.hooksPath hooks

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: coverage-html
coverage-html: ## Generate test coverage report and open in browser
	go test ./... -coverpkg=./... -coverprofile=test-coverage.out
	go tool cover -html=test-coverage.out

.PHONY: lint
lint: ## Verify code style and run static checks
	go vet -asmdecl -assign -atomic -bools -buildtag -cgocall -copylocks -httpresponse -loopclosure -lostcancel -nilfunc -printf -shift -stdmethods -structtag -tests -unmarshal -unreachable -unsafeptr -unusedresult ./...
	test -z $(gofmt -l ./...)

.PHONY: fmt
fmt: ## Run syntax re-formatting (modify in place)
	go fmt ./...

.PHONY: check
check: ## Compile everything, checking syntax (does not output binaries)
	go build ./...

.PHONY: env
.env:
	if [ ! -f ".env" ]; then cp example.dev.env .env; fi


.PHONY: submodules
submodules:
	git submodule foreach git pull origin main
