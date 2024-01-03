#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/clean.sh
${SCRIPTS}/generate.sh
${SCRIPTS}/docs.sh

mkdir -p ${TOOLS_DIR}

# Ensure all source files have copyright header.
go install github.com/google/addlicense@v1.1.1
${TOOLS_DIR}/addlicense -f license.tpl \
    -l mpl -c XigXog -y 2023 \
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
