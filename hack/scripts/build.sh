#!/bin/bash

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

export GO111MODULE=on
export CGO_ENABLED=0

BUILD_DATE=$(TZ=UTC date --iso-8601=seconds)

if ! ${SKIP_GENERATE:-false}; then
	$SCRIPTS/generate.sh
fi

mkdir -p "${BUILD_OUT}"
rm -f "${COMPONENT_OUT}"

go build \
	-C "${COMPONENT_SRC}/" -o "${COMPONENT_OUT}" \
	-ldflags " \
		-X github.com/xigxog/kubefox/core.BuildDate=${BUILD_DATE}\
		-X github.com/xigxog/kubefox/core.Component=${COMPONENT} \
		-X github.com/xigxog/kubefox/core.Commit=${COMPONENT_COMMIT} \
		-X github.com/xigxog/kubefox/core.RootCommit=${ROOT_COMMIT} \
		-X github.com/xigxog/kubefox/core.HeadRef=${HEAD_REF} \
		-X github.com/xigxog/kubefox/core.TagRef=${TAG_REF}" \
	main.go
