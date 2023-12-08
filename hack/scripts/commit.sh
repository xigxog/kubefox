#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/clean.sh
${SCRIPTS}/generate.sh
${SCRIPTS}/docs.sh

gofmt -l -s -w .
go vet ./...

git add .
git commit
