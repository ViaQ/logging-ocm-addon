VERSION ?= 0.0.1

# REGISTRY_BASE
# defines the container registry and organization for the bundle and operator container images.
REGISTRY_BASE_OPENSHIFT = quay.io/openshift-logging
REGISTRY_BASE ?= $(REGISTRY_BASE_OPENSHIFT)

# Image URL to use all building/pushing image targets
IMG ?= $(REGISTRY_BASE)/logging-omc-addon:$(VERSION)

.PHONY: deps
deps: go.mod go.sum
	go mod tidy
	go mod download
	go mod verify

.PHONY: addon
addon: deps ## Build addon binary
	go build -o bin/logging-omc-addon cmd/helm/main.go

.PHONY: oci-build
oci-build: ## Build the image
	podman build -t ${IMG} .

.PHONY: oci-push
oci-push: ## Push the image
	podman push ${IMG}

.PHONY: oci
oci: oci-build oci-push