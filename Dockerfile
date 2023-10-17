# FROM ghcr.io/xigxog/base
FROM alpine
ARG COMPONENT
COPY ./bin/${COMPONENT} /component
ENTRYPOINT [ "/component" ]
