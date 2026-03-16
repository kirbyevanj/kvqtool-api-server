FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY kvq-models/ /kvq-models/

COPY api-server/go.mod api-server/go.sum ./
RUN go mod edit -replace github.com/kirbyevanj/kvqtool-kvq-models=/kvq-models \
    && go mod download

COPY api-server/ .
RUN CGO_ENABLED=0 go build -o /api-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /api-server /api-server
EXPOSE 8080
ENTRYPOINT ["/api-server"]
