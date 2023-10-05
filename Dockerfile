FROM gcr.io/distroless/static:nonroot
ARG COMPONENT

COPY ./bin/${COMPONENT} /component
ENTRYPOINT [ "/component" ]
