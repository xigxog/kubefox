# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

ARG COMPONENT

## Builder
FROM ghcr.io/xigxog/builder:v0.1.0 AS builder

ARG COMPONENT
ARG TARGETPLATFORM
ARG COMPRESS=false

ENV ARCH=${TARGETPLATFORM##*/}

WORKDIR /workspace

RUN go env -w GOCACHE=/go-cache && \
    go env -w GOMODCACHE=/gomod-cache && \
    go env -w CGO_ENABLED=0

COPY go.mo[d] go.su[m] ./
RUN --mount=type=cache,target=/gomod-cache \
    go mod download

COPY ./ ./
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    make build COMPONENT=${COMPONENT}

RUN if ${COMPRESS}; then upx ./bin/${COMPONENT}; fi

## Runtime
FROM ghcr.io/xigxog/base:v0.2.0

ARG COMPONENT

COPY --from=builder /workspace/bin/${COMPONENT} /component

ENTRYPOINT [ "/component" ]
