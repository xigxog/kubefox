#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

pipenv install
pipenv run mkdocs serve
