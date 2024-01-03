#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/clean.sh
${SCRIPTS}/generate.sh
${SCRIPTS}/docs.sh

mkdir -p ${TOOLS_DIR}

# Ensure all source files have copyright header.
go install github.com/google/addlicense@v1.1.1
${TOOLS_DIR}/addlicense -f hack/license.tpl \
    -l mpl -c XigXog -y 2023 \
    -ignore ".markdownlint.yaml" \
    -ignore ".github/**" \
    -ignore "examples/**" \
    -ignore "site/**" \
    -ignore "workstation/**" \
    .

go mod tidy
gofmt -l -s -w .
go vet ./...

git add .
git commit
