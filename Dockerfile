# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

# Compress binary
FROM ghcr.io/xigxog/upx:4.2.1 AS upx

ARG COMPONENT
ARG COMPRESS=false

COPY ./bin/${COMPONENT} /component
RUN if ${COMPRESS}; then upx /component; fi

# Runtime
FROM ghcr.io/xigxog/base:v0.2.0
COPY --from=upx /component /component
ENTRYPOINT [ "/component" ]
