FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /auth-service \
    ./cmd/auth_service


FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8081

CMD ["/app/auth-service"]