ARG GOLANG_VERSION="1.23.3"
ARG PROJECT="updown-webhook"

ARG TARGETOS
ARG TARGETARCH

ARG COMMIT
ARG VERSION

FROM --platform=${TARGETARCH} golang:${GOLANG_VERSION} AS build

ARG PROJECT

WORKDIR /${PROJECT}

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/webhook cmd/webhook
COPY updown updown
COPY webhook webhook

ARG TARGETOS
ARG TARGETARCH

ARG COMMIT
ARG VERSION

RUN BUILD_TIME=$(date +%s) && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -a \
    -installsuffix cgo \
    -ldflags "-X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${COMMIT}' -X 'main.OSVersion=${VERSION}'" \
    -o /bin/webhook \
    ./cmd/webhook


FROM --platform=${TARGETARCH} gcr.io/distroless/static-debian12:latest

LABEL org.opencontainers.image.source="https://github.com/DazWilkin/updown-webhook"

ARG PROJECT

COPY --from=build /bin/webhook /bin/webhook

ENTRYPOINT ["/bin/webhook"]
CMD ["--port=8888"]
