FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod init ms-wishlist || true
RUN go mod tidy || true

RUN CGO_ENABLED=0 GOOS=linux go build -o api-wishlist .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-wishlist .

EXPOSE 8082

CMD ["./api-wishlist"]