FROM alpine

ARG component
ENV KUBEFOX_COMPONENT=${component}

COPY ./bin/${KUBEFOX_COMPONENT} .
RUN mv ${KUBEFOX_COMPONENT} component

ENTRYPOINT [ "./component" ]