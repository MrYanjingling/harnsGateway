# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
ARG SERVICE=platform
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app
COPY --from=builder /out/app /usr/local/bin/app
USER app
ENTRYPOINT ["app"]
