#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

CONTROLLER_GEN_VERSION="v0.15.0"
PROTOC_GEN_GO_VERSION="v1.28"
PROTOC_GEN_GO_GRPC_VERSION="1.2.0" # --version does not output 'v'

rm -rf ${CRDS_OUT}
mkdir -p ${TOOLS_DIR} ${CRDS_OUT}

### Install tools. If wrong version is installed, tool will be overwritten.
${TOOLS_DIR}/controller-gen --version | grep -q ${CONTROLLER_GEN_VERSION} ||
    go install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}

${TOOLS_DIR}/protoc-gen-go --version | grep -q ${PROTOC_GEN_GO_VERSION} ||
    go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}

${TOOLS_DIR}/protoc-gen-go-grpc --version | grep -q ${PROTOC_GEN_GO_GRPC_VERSION} ||
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v${PROTOC_GEN_GO_GRPC_VERSION}
###

# Generate code containing DeepCopy, DeepCopyInto, DeepCopyObject and CRDs.
${TOOLS_DIR}/controller-gen paths="{./${API_DIR}/...}" \
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

${SCRIPTS}/addlicense.sh

gofmt -l -s -w .
