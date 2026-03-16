FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
ENV GOTOOLCHAIN=auto
ENV GONOSUMCHECK=github.com/kirbyevanj/*
ENV GOPRIVATE=github.com/kirbyevanj/*
ENV GOFLAGS=-mod=mod

COPY kvq-models/ /kvq-models/
COPY api-server/ .

RUN go mod edit -replace github.com/kirbyevanj/kvqtool-kvq-models=/kvq-models \
    && go mod tidy \
    && CGO_ENABLED=0 go build -o /api-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /api-server /api-server
EXPOSE 8080
ENTRYPOINT ["/api-server"]
