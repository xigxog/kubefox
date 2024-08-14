#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

rm -rf ${CRDS_OUT}
mkdir -p ${CRDS_OUT}

# Generate code containing DeepCopy, DeepCopyInto, DeepCopyObject and CRDs.
controller-gen paths="{./${API_DIR}/...}" \
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
