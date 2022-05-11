FROM gcr.io/distroless/static
LABEL key="value"

ARG NAME

COPY ${NAME} /server

ENTRYPOINT ["/server"]
