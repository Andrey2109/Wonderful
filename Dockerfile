FROM golang:1.25.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o cli-app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cli-app .
COPY --from=builder /app/instructions.txt .
ENTRYPOINT ["/app/cli-app"]