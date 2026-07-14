FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /vpn-manager ./cmd/vpn-manager

FROM alpine:3.19

RUN apk add --no-cache ca-certificates
RUN mkdir -p /data

COPY --from=builder /vpn-manager /usr/local/bin/vpn-manager

EXPOSE 8080

ENTRYPOINT ["vpn-manager"]
CMD ["-config", "/etc/vpn-manager/config.json"]
