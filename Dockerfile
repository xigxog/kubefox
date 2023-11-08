# Compress binary
FROM ghcr.io/xigxog/upx:4.2.0 AS upx

ARG COMPONENT
ARG COMPRESS=false

COPY ./bin/${COMPONENT} /component
RUN if ${COMPRESS}; then upx /component; fi

# Runtime
FROM ghcr.io/xigxog/base:v0.2.0
COPY --from=upx /component /component
ENTRYPOINT [ "/component" ]
