# Note recursive assignment. The expression is evaluated at usage to ensure
# $(component) has been set.
SRC_DIR = $(abspath $(ROOT)/components/$(component))
IMAGE = $(IMAGE_REGISTRY)/kubefox/$(component)

IMAGE_TAG := latest
IMAGE_REGISTRY := ghcr.io/xigxog

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TARGET_DIR := $(abspath $(ROOT)/bin)

API_DIR := $(abspath $(ROOT)/libs/core/api)
CRDS_DIR := $(abspath $(ROOT)/libs/core/api/crds)
CRDS_BOOTSTRAP := $(CRDS_DIR)/bootstrap
PROTO_SRC := $(abspath $(ROOT)/libs/core/api/protobuf)
PROTO_OUT := $(abspath $(ROOT)/libs/core/grpc)
DOCS_OUT := $(abspath $(ROOT)/site)

COMPONENTS := api-server broker operator runtime-server
PUSHES := $(addprefix push/,$(COMPONENTS))
IMAGES := $(addprefix image/,$(COMPONENTS))
BINS := $(addprefix bin/,$(COMPONENTS))
INSPECTS := $(addprefix inspect/,$(COMPONENTS))


.PHONY: all
all: clean generate $(BINS)

.PHONY: push-all
push-all: clean generate $(PUSHES)

.PHONY: image-all
image-all: clean generate $(IMAGES)

.PHONY: $(PUSHES)
$(PUSHES):
	$(eval component=$(notdir $@))
	$(eval git_hash=$(shell git log -n 1 --format="%h" -- components/$(component)/))
	$(MAKE) "image/$(component)"

	buildah push "$(IMAGE):$(git_hash)"
	buildah push "$(IMAGE):$(IMAGE_TAG)"

.PHONY: $(IMAGES)
$(IMAGES):
	$(eval component=$(notdir $@))
	$(eval git_hash=$(shell git log -n 1 --format="%h" -- components/$(component)/))
	$(eval container=$(shell buildah from gcr.io/distroless/static))
	$(MAKE) bin/$(component)

	buildah add $(container) "$(TARGET_DIR)/$(component)"
	buildah config --entrypoint '["./$(component)"]' $(container) 
	buildah commit $(container) "$(IMAGE):$(git_hash)"
	buildah tag "$(IMAGE):$(git_hash)" "$(IMAGE):$(IMAGE_TAG)"

.PHONY: $(BINS)
$(BINS):
	$(eval component=$(notdir $@))

	mkdir -p "$(dir $@)"
	CGO_ENABLED=0 go build -C "$(SRC_DIR)/" -o "$(TARGET_DIR)/$(component)" -ldflags "-s -w" main.go

.PHONY: docs
docs: clean generate
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1
	protoc \
		--proto_path=$(PROTO_SRC) \
		--doc_out=docs/reference/ --doc_opt=docs/protobuf.tmpl,protobuf.md \
		kubefox.proto

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
		--go_out=$(PROTO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=paths=source_relative \
		kubefox.proto

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
.PHONY: crds
crds:
	mkdir -p $(CRDS_DIR)
	@# A file needs to be present in dir for api/embed.go to compile
	touch $(CRDS_BOOTSTRAP)
	$(ROOT)/utils/controller-gen paths="$(API_DIR)/..." object crd output:crd:artifacts:config="$(CRDS_DIR)"
	rm -f $(CRDS_BOOTSTRAP)

.PHONY: clean
clean:
	rm -rf $(TARGET_DIR)/
	rm -rf $(CRDS_DIR)/
	rm -rf $(DOCS_OUT)/
