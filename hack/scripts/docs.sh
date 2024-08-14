#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

FOX_ROOT="${REPO_ROOT}/../fox"

CRD_DOCS="${DOCS_SRC}/reference/kubernetes-crds.md"
KIT_DOCS_GO="${DOCS_SRC}/reference/kit/go"

rm -rf ${DOCS_OUT}

# Generate docs from proto files.
protoc \
    --proto_path=./${PROTO_SRC} \
    --doc_out=./${DOCS_SRC}/reference/ --doc_opt=./${DOCS_SRC}/templates/protobuf/protobuf.tmpl,protobuf.md \
    protobuf_msgs.proto broker_svc.proto \
    google/protobuf/struct.proto \
    opentelemetry/proto/common/v1/common.proto \
    opentelemetry/proto/logs/v1/logs.proto \
    opentelemetry/proto/metrics/v1/metrics.proto \
    opentelemetry/proto/resource/v1/resource.proto \
    opentelemetry/proto/trace/v1/trace.proto

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

crd-ref-docs \
    --config ${DOCS_SRC}/crd-ref-docs.yaml \
    --source-path ${REPO_ROOT}/api/ \
    --output-path ${CRD_DOCS} \
    --templates-dir ${DOCS_SRC}/templates/crd-ref-docs \
    --renderer markdown

# Generate docs from source files.
rm -rf "${KIT_DOCS_GO}"
mkdir -p "${KIT_DOCS_GO}"

find "${REPO_ROOT}/kit" -type d -exec bash -c '
echo "generating go docs for ${1}" && \
    gomarkdoc --output "${0}/$(basename "${1}").md" "${1}"
' "${KIT_DOCS_GO}" {} \;

pipenv install
pipenv run mkdocs build
