#!/bin/bash

set -o errexit

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." &>/dev/null && pwd -P)"
cd "${REPO_ROOT}"

SCRIPTS="hack/scripts"

API_DIR="api"
COMPONENTS_DIR="components"
CORE_DIR="core"
TOOLS_DIR="tools"

COMPONENT_SRC="${COMPONENTS_DIR}/${COMPONENT}"
DOCS_SRC="docs"
PROTO_SRC="${API_DIR}/protobuf"

BUILD_OUT="bin"
COMPONENT_OUT=${COMPONENT_OUT:-"${REPO_ROOT}/${BUILD_OUT}/${COMPONENT}"}
CRDS_OUT="${API_DIR}/crds"
DOCS_OUT="site"
GRPC_OUT="grpc"
PROTO_OUT="core"

COMPONENT_COMMIT=$(git log -n 1 --format="%H" -- ${COMPONENT_SRC}/)
ROOT_COMMIT=$(git rev-parse HEAD)

HEAD_REF=$(git symbolic-ref -q HEAD)
TAG_REF=$(git describe --tags --exact-match 2>/dev/null | xargs -I % echo "refs/tags/%")

IMAGE_TAG=${IMAGE_TAG:-$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match)}
IMAGE="ghcr.io/xigxog/kubefox/${COMPONENT}:${IMAGE_TAG}"

set -o pipefail -o xtrace -o nounset
