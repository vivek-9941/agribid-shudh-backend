# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /agribid-server ./cmd/server

# --- Final stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /agribid-server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/internal/invoice/templates ./internal/invoice/templates

EXPOSE 8080 9090

ENTRYPOINT ["/app/agribid-server"]
