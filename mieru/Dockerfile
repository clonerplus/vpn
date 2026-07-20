FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
RUN git clone --depth 1 https://github.com/enfein/mieru.git . && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o mita cmd/mita/mita.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -H -D -g "" mita && \
    mkdir -p /etc/mita /var/lib/mita /var/run/mita && \
    chown -R mita:mita /etc/mita /var/lib/mita /var/run/mita

COPY --from=builder /build/mita /usr/local/bin/mita

USER mita
CMD ["/usr/local/bin/mita", "run"]
