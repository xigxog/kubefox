#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

PUSH_ARG=""
if ${PUSH}; then
	PUSH_ARG="--push"
fi

DOCKERFILE_ARG=""
if ${DEBUG}; then
	DOCKERFILE_ARG="--file Dockerfile.debug ."
fi

docker buildx build \
	--build-arg COMPONENT="${COMPONENT}" \
	--build-arg COMPRESS="${COMPRESS}" \
	--platform linux/amd64,linux/arm64 \
	--tag "${IMAGE}" \
	--progress plain \
	${DOCKERFILE_ARG} \
	${PUSH_ARG} \
	.

if ${KIND_LOAD}; then
	kind load docker-image --name "${KIND_NAME}" "${IMAGE}"
fi
