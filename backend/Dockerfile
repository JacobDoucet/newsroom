FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o newsroom-api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/newsroom-api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./newsroom-api"]
