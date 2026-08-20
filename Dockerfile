FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /app/bin/shorty ./cmd/shorty

FROM alpine:latest

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/bin/shorty .

COPY config.json .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["./shorty", "-c", "./config.json"]