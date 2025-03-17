FROM alpine:3.18

ARG NAME

RUN apk --no-cache add ca-certificates && \
    update-ca-certificates

COPY ${NAME} /app

ENTRYPOINT ["/app"]
