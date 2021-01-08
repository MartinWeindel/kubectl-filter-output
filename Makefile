BINARY_NAME := kubectl-fo

.PHONY: install
install: ## Install the binary 
	@CGO_ENABLED=0  go build -o ${GOPATH}/bin/${BINARY_NAME} -ldflags="-s -w"