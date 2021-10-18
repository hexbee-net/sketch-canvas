# syntax=docker/dockerfile:1
FROM golang:1.17 AS builder
WORKDIR /go/src/github.com/hexbee-net/sketch-canvas

COPY . .

# Fetch dependencies
RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo ./cmd/server

FROM scratch

# copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /bin/

COPY --from=builder /go/src/github.com/hexbee-net/sketch-canvas/server .

ENTRYPOINT [ "/bin/server" ]
EXPOSE 8800
