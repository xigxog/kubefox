# Note recursive assignment. The expression is evaluated at usage to ensure
# $(component) has been set.
SRC_DIR = $(abspath $(ROOT)/components/$(component))
GIT_COMMIT = $(shell git log -n 1 --format="%h" -- components/$(component)/)
IMAGE = $(IMAGE_REGISTRY)/kubefox/$(component)

GIT_REF := $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
IMAGE_REGISTRY := ghcr.io/xigxog
KIND_NAME := kubefox
KIND_LOAD := false

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TARGET_DIR := $(abspath $(ROOT)/bin)
API_DIR := $(abspath $(ROOT)/libs/core/api)
CRDS_DIR := $(abspath $(API_DIR)/crds)
PROTO_SRC := $(abspath $(API_DIR)/protobuf)
KUBEFOX_DIR := $(abspath $(ROOT)/libs/core/kubefox)
GRPC_OUT := $(abspath $(ROOT)/libs/core/grpc)
DOCS_OUT := $(abspath $(ROOT)/site)

TOOLS_DIR := $(abspath $(ROOT)/tools)
CONTROLLER_GEN := $(TOOLS_DIR)/controller-gen
CONTROLLER_TOOLS_VERSION := v0.13.0

COMPONENTS := broker operator
PUSHES := $(addprefix push/,$(COMPONENTS))
IMAGES := $(addprefix image/,$(COMPONENTS))
BINS := $(addprefix bin/,$(COMPONENTS))
TIDIES := $(addprefix tidy/,$(COMPONENTS))
INSPECTS := $(addprefix inspect/,$(COMPONENTS))


.PHONY: all
all: clean generate $(BINS)

.PHONY: push-all
push-all: clean generate $(PUSHES)

.PHONY: image-all
image-all: clean generate $(IMAGES)

.PHONY: tidy-all
tidy-all: $(TIDIES)

.PHONY: $(PUSHES)
$(PUSHES):
	$(eval component=$(notdir $@))
	$(MAKE) "image/$(component)"

	buildah push "$(IMAGE):$(GIT_REF)"

.PHONY: $(IMAGES)
$(IMAGES):
	$(eval component=$(notdir $@))
	$(MAKE) bin/$(component)

	buildah bud --build-arg COMPONENT="$(component)" --tag "$(IMAGE):$(GIT_REF)"
	if $(KIND_LOAD); then \
		buildah push "$(IMAGE):$(GIT_REF)" "docker-daemon:$(IMAGE):$(GIT_REF)"; \
		kind load docker-image --name "$(KIND_NAME)" "$(IMAGE):$(GIT_REF)"; \
	fi

.PHONY: $(BINS)
$(BINS):
	$(eval component=$(notdir $@))

	mkdir -p "$(dir $@)"
	CGO_ENABLED=0 go build \
		-C "$(SRC_DIR)/" -o "$(TARGET_DIR)/$(component)" \
		-ldflags "-X main.GitRef=$(GIT_REF) -X main.GitCommit=$(GIT_COMMIT)" \
		main.go

.PHONY: $(TIDIES)
$(TIDIES):
	$(eval component=$(notdir $@))

	cd components/$(component) && go mod tidy

.PHONY: docs
docs:
	rm -rf $(DOCS_OUT)/
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1
	protoc \
		--proto_path=$(PROTO_SRC) \
		--doc_out=docs/reference/ --doc_opt=docs/protobuf.tmpl,protobuf_msgs.md \
		protobuf_msgs.proto

	protoc \
		--proto_path=$(PROTO_SRC) \
		--doc_out=docs/reference/ --doc_opt=docs/protobuf.tmpl,broker_svc.md \
		broker_svc.proto

	pipenv install
	pipenv run mkdocs build

.PHONY: serve-docs
serve-docs:
	pipenv install
	pipenv run mkdocs serve

.PHONY: generate
generate: protobuf crds

.PHONY: protobuf
protobuf:
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(KUBEFOX_DIR) \
		--go_opt=paths=source_relative \
		protobuf_msgs.proto
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go-grpc_out=$(GRPC_OUT) \
		--go-grpc_opt=paths=source_relative \
		broker_svc.proto

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
.PHONY: crds
crds: controller-gen
	mkdir -p $(CRDS_DIR)
	$(CONTROLLER_GEN) paths="{$(KUBEFOX_DIR)/..., $(API_DIR)/kubernetes/...}" object crd output:crd:artifacts:config="$(CRDS_DIR)"

.PHONY: clean
clean:
	rm -rf $(TARGET_DIR)/
	rm -rf $(CRDS_DIR)/
	rm -rf $(DOCS_OUT)/
	rm -rf $(TOOLS_DIR)/

# Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
.PHONY: controller-gen
controller-gen:
	mkdir -p $(TOOLS_DIR)
	test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(TOOLS_DIR) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
