# Сборка
FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go clean --modcache
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -v -o sub_service cmd/app/main.go

FROM alpine:3.22
WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/sub_service ./sub_service
COPY --from=builder /app/.env ./.env
RUN chown -R appuser:appgroup /app && chmod +x /app/sub_service
USER appuser
EXPOSE 8080
CMD ["./sub_service"]