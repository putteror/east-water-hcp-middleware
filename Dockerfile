# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server cmd/app/main.go

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Bangkok

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env

EXPOSE 8080

CMD ["./server"]
