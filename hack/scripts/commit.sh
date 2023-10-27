#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

$SCRIPTS/clean.sh
$SCRIPTS/generate.sh

go fmt ./...
go vet ./...

git add .
git commit
