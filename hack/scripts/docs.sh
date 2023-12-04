#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

PROTOC_GEN_DOC_VERSION="v1.5.1"
FOX_ROOT="${REPO_ROOT}/../fox"

rm -rf ${DOCS_OUT}

### Install tools. If wrong version is installed, tool will be overwritten.
${TOOLS_DIR}/protoc-gen-doc -version | grep -q ${PROTOC_GEN_DOC_VERSION} ||
    go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@${PROTOC_GEN_DOC_VERSION}
###

# Generate docs from proto files.
protoc \
    --proto_path=./${PROTO_SRC} \
    --doc_out=./${DOCS_SRC}/reference/ --doc_opt=./${DOCS_SRC}/protobuf.tmpl,protobuf.md \
    protobuf_msgs.proto broker_svc.proto

if [ -d "${FOX_ROOT}" ]; then
    (
        cd "${FOX_ROOT}"
        make docs
    )
    rm -rf ${DOCS_SRC}/fox
    cp -r ${FOX_ROOT}/docs ${DOCS_SRC}/fox
fi

pipenv install
pipenv run mkdocs build
