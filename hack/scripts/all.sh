#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0


source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

export SKIP_GENERATE=true

${SCRIPTS}/generate.sh

for dir in ${COMPONENTS_DIR}/*/; do
    export COMPONENT="$(basename ${dir})"
    ${SCRIPTS}/image.sh &
done

wait
