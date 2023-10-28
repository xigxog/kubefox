#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

export SKIP_GENERATE=true

${SCRIPTS}/generate.sh

for dir in ${COMPONENTS_DIR}/*/; do
    export COMPONENT="$(basename ${dir})"
    ${SCRIPTS}/image.sh &
done

wait
