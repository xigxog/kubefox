#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

PROTOC_GEN_DOC_VERSION="v1.5.1"
GOMARKDOC_VERSION="v1.1.0"
CRD_REF_DOCS_VERSION="v0.1.5"

FOX_ROOT="${REPO_ROOT}/../fox"

CRD_DOCS="${DOCS_SRC}/reference/kubernetes-crds.md"
KIT_DOCS_GO="${DOCS_SRC}/reference/kit/go"

rm -rf ${DOCS_OUT}

### Install tools. If wrong version is installed, tool will be overwritten.
${TOOLS_DIR}/protoc-gen-doc -version | grep -q ${PROTOC_GEN_DOC_VERSION} ||
    go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@${PROTOC_GEN_DOC_VERSION}

${TOOLS_DIR}/gomarkdoc --version | grep -q ${GOMARKDOC_VERSION} ||
    go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@${GOMARKDOC_VERSION}

go install github.com/xigxog/crd-ref-docs@${CRD_REF_DOCS_VERSION}
###

# Generate docs from proto files.
protoc \
    --proto_path=./${PROTO_SRC} \
    --doc_out=./${DOCS_SRC}/reference/ --doc_opt=./${DOCS_SRC}/templates/protobuf/protobuf.tmpl,protobuf.md \
    protobuf_msgs.proto broker_svc.proto

if [ -d "${FOX_ROOT}" ]; then
    (
        cd "${FOX_ROOT}"
        make docs
    )
    rm -rf ${DOCS_SRC}/reference/fox
    cp -r ${FOX_ROOT}/docs ${DOCS_SRC}/reference/fox
fi

# Generate docs from CRDs.
rm -f "${CRD_DOCS}"
mkdir -p $(dirname "${CRD_DOCS}")

${TOOLS_DIR}/crd-ref-docs \
    --config ${DOCS_SRC}/crd-ref-docs.yaml \
    --source-path ${REPO_ROOT}/api/ \
    --output-path ${CRD_DOCS} \
    --templates-dir ${DOCS_SRC}/templates/crd-ref-docs \
    --renderer markdown

# Generate docs from source files.
rm -rf "${KIT_DOCS_GO}"
mkdir -p "${KIT_DOCS_GO}"

find "${REPO_ROOT}/kit" -type d -exec bash -c '
echo "generating go docs for ${2}" && \
    ${0}/gomarkdoc --output "${1}/$(basename "${2}").md" "${2}"
' "${TOOLS_DIR}" "${KIT_DOCS_GO}" {} \;

pipenv install
pipenv run mkdocs build
