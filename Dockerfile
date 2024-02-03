FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/raft .

# Path: Dockerfile
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/raft .


CMD ["./raft"]