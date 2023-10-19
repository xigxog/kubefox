FROM ghcr.io/xigxog/base
ARG COMPONENT
COPY ./bin/${COMPONENT} /component
ENTRYPOINT [ "/component" ]
