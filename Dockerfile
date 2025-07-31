FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o vedic-astrology-server main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/vedic-astrology-server .
COPY --from=builder /app/post-prompt.txt .

EXPOSE 9494

CMD ["./vedic-astrology-server"]