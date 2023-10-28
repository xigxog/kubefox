#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

${SCRIPTS}/build.sh

if ${DEBUG}; then
	DOCKERFILE="--file Dockerfile.debug ."
fi

buildah bud --build-arg COMPONENT="${COMPONENT}" --build-arg COMPRESS="${COMPRESS}" --tag "${IMAGE}" ${DOCKERFILE}
if ${KIND_LOAD}; then
	buildah push "${IMAGE}" "docker-daemon:${IMAGE}"
	kind load docker-image --name "${KIND_NAME}" "${IMAGE}"
fi

if ${PUSH}; then
	buildah push "${IMAGE}"
fi
