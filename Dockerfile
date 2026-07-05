FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/vpn-server ./cmd/vpn-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/vpn-server /usr/local/bin/vpn-server

RUN adduser -D -h /home/vpn vpn
USER vpn

EXPOSE 443 8080 51820/udp

ENTRYPOINT ["vpn-server"]
