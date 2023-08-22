
.PHONY: deps
deps: go.mod go.sum
	go mod tidy
	go mod download
	go mod verify

.PHONY: helm-agent
helm-agent: deps ## Build helm-agent binary
	go build -o bin/helm-agent cmd/helm/main.go