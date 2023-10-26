#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

SKIP_GENERATE=true
RELEASE=${RELEASE:-false}

$SCRIPTS/generate.sh

for dir in ${COMPONENTS_DIR}/*/; do
    export COMPONENT="$(basename ${dir})"

    if ${RELEASE}; then
        $SCRIPTS/push.sh
    else
        $SCRIPTS/image.sh
    fi
done
