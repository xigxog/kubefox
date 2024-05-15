#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

set -o xtrace
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
BROKER_COMMIT=$(git log -n 1 --format="%H" -- ${COMPONENTS_DIR}/broker/)
HTTPSRV_COMMIT=$(git log -n 1 --format="%H" -- ${COMPONENTS_DIR}/httpsrv/)
OPERATOR_COMMIT=$(git log -n 1 --format="%H" -- ${COMPONENTS_DIR}/operator/)
ROOT_COMMIT=$(git rev-parse HEAD)
VERSION=${VERSION:""}

HEAD_REF=$(git symbolic-ref -q HEAD || true)
TAG_REF=$(git describe --tags --exact-match 2>/dev/null | xargs -I % echo "refs/tags/%")

CONTAINER_REGISTRY=${CONTAINER_REGISTRY:-"ghcr.io/xigxog"}
IMAGE_TAG=${IMAGE_TAG:-$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match)}
IMAGE=${IMAGE:-"${CONTAINER_REGISTRY}/kubefox/${COMPONENT}:${IMAGE_TAG}"}

export GO111MODULE=on
export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=linux
export GOBIN="${REPO_ROOT}/${TOOLS_DIR}"
export PATH="${PATH}:${GOBIN}"

BUILD_DATE=$(TZ=UTC date --iso-8601=seconds)

PUSH=${PUSH:-false}
COMPRESS=${COMPRESS:-false}
DEBUG=${DEBUG:-false}
SKIP_GENERATE=${SKIP_GENERATE:-false}

KIND_NAME=${KIND_NAME:-"kind"}
KIND_LOAD=${KIND_LOAD:-false}
DOCKERFILE=""

set -o pipefail -o nounset
