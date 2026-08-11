FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=local
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o catalogsync ./cmd/catalogsync && \
    CGO_ENABLED=0 GOOS=linux go build -o import-wanted ./cmd/import-wanted

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/catalogsync .
COPY --from=builder /app/import-wanted .
COPY seeds ./seeds
RUN mkdir -p /app/uploads

EXPOSE 8080
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0
ENV ENVIRONMENT=production

# Миграции вшиты в бинарь и накатываются при старте (AUTO_MIGRATE=false отключает).
CMD ["./server"]
