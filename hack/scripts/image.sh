#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

KIND_NAME=${KIND_NAME:-"kind"}
KIND_LOAD=${KIND_LOAD:-"false"}

$SCRIPTS/build.sh

buildah bud --build-arg COMPONENT="${COMPONENT}" --tag "${IMAGE}"
if ${KIND_LOAD}; then
	buildah push "${IMAGE}" "docker-daemon:${IMAGE}"
	kind load docker-image --name "${KIND_NAME}" "${IMAGE}"
fi
