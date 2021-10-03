FROM golang
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build
ENTRYPOINT /app/phs

# resulting container image
FROM alpine:latest
RUN apk --no-cache add ca-certificates && \
    mkdir /config/
WORKDIR /
COPY --from=0 /app/phs /
CMD ["/phs"]

