# Stage 1: Builder
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /build/server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata postgresql-client

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/server /app/tutoring-backend
COPY --from=builder /build/internal/database/migrations /app/migrations
COPY entrypoint.sh /app/entrypoint.sh

RUN mkdir -p /app/uploads && \
    chown -R appuser:appgroup /app && \
    chmod -R u+rwX /app && \
    chmod +x /app/entrypoint.sh /app/tutoring-backend

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
