#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

CONTROLLER_GEN_VERSION="v0.13.0"
PROTOC_GEN_DOC_VERSION="v1.5.1"

rm -rf ${CRDS_OUT}
mkdir -p ${TOOLS_DIR} ${CRDS_OUT}

# Download controller-gen and protoc-gen-doc. If wrong version is installed,
# it will be overwritten.
${TOOLS_DIR}/controller-gen --version | grep -q ${CONTROLLER_GEN_VERSION} ||
    GOBIN="${REPO_ROOT}/${TOOLS_DIR}" go install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}

${TOOLS_DIR}/protoc-gen-doc -version | grep -q ${PROTOC_GEN_DOC_VERSION} ||
    GOBIN="${REPO_ROOT}/${TOOLS_DIR}" go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@${PROTOC_GEN_DOC_VERSION}

# Generate code containing DeepCopy, DeepCopyInto, DeepCopyObject and CRDs.
${TOOLS_DIR}/controller-gen paths="{./${CORE_DIR}/..., ./${API_DIR}/kubernetes/...}" \
    object crd output:crd:artifacts:config=./${CRDS_OUT}/

# Generate code from proto files.
protoc \
    --proto_path=./${PROTO_SRC} \
    --go_out=./${PROTO_OUT} \
    --go_opt=paths=source_relative \
    protobuf_msgs.proto
protoc \
    --proto_path=./${PROTO_SRC} \
    --go-grpc_out=./${GRPC_OUT} \
    --go-grpc_opt=paths=source_relative \
    broker_svc.proto

# Generate docs from proto files.
protoc \
    --plugin=protoc-gen-doc=${TOOLS_DIR}/protoc-gen-doc \
    --proto_path=./${PROTO_SRC} \
    --doc_out=./${DOCS_SRC}/reference/ --doc_opt=./${DOCS_SRC}/protobuf.tmpl,protobuf.md \
    protobuf_msgs.proto broker_svc.proto
