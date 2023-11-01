#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

if ! ${SKIP_GENERATE}; then
	${SCRIPTS}/generate.sh
fi

mkdir -p "${BUILD_OUT}"
rm -f "${COMPONENT_OUT}"

go build \
	-C "${COMPONENT_SRC}/" -o "${COMPONENT_OUT}" \
	-ldflags " \
		-w -s
		-X github.com/xigxog/kubefox/build.date=${BUILD_DATE}\
		-X github.com/xigxog/kubefox/build.component=${COMPONENT} \
		-X github.com/xigxog/kubefox/build.commit=${COMPONENT_COMMIT} \
		-X github.com/xigxog/kubefox/build.rootCommit=${ROOT_COMMIT} \
		-X github.com/xigxog/kubefox/build.headRef=${HEAD_REF} \
		-X github.com/xigxog/kubefox/build.tagRef=${TAG_REF}" \
