#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

rm -rf ${BUILD_OUT} ${CRDS_OUT} ${DOCS_OUT} ${TOOLS_DIR}
