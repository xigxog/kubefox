#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0


source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/build.sh

if ${DEBUG}; then
	DOCKERFILE="--file Dockerfile.debug ."
fi

# Required dependency: https://github.com/containers/buildah/blob/main/install.md
buildah bud --build-arg COMPONENT="${COMPONENT}" --build-arg COMPRESS="${COMPRESS}" --tag "${IMAGE}" ${DOCKERFILE}
if ${KIND_LOAD}; then
	buildah push "${IMAGE}" "docker-daemon:${IMAGE}"
	kind load docker-image --name "${KIND_NAME}" "${IMAGE}"
fi

if ${PUSH}; then
	buildah push "${IMAGE}"
fi
