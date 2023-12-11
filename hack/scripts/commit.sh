#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/clean.sh
${SCRIPTS}/generate.sh
${SCRIPTS}/docs.sh

go mod tidy
gofmt -l -s -w .
go vet ./...

git add .
git commit
