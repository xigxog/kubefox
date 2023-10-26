#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

rm -rf ${DOCS_OUT}

pipenv install
pipenv run mkdocs build
