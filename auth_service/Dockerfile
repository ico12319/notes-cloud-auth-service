# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /auth-service \
    ./cmd/auth_service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8081

CMD ["/app/auth-service"]