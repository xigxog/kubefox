REPO_ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all
all:
	$(REPO_ROOT)hack/scripts/all.sh

.PHONY: generate
generate:
	$(REPO_ROOT)hack/scripts/generate.sh

.PHONY: build
build:
	$(REPO_ROOT)hack/scripts/build.sh

.PHONY: image
image:
	$(REPO_ROOT)hack/scripts/image.sh

.PHONY: docs
docs:
	$(REPO_ROOT)hack/scripts/docs.sh

.PHONY: serve-docs
serve-docs:
	$(REPO_ROOT)hack/scripts/serve-docs.sh

.PHONY: clean
clean:
	$(REPO_ROOT)hack/scripts/clean.sh

.PHONY: commit
commit:
	$(REPO_ROOT)hack/scripts/commit.sh
