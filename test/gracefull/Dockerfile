FROM golang:1.24.5-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/agent ./cmd/agent
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim AS agent
RUN apt-get update -qq && apt-get install -y -qq ca-certificates >/dev/null && update-ca-certificates
COPY --from=builder /out/agent /usr/local/bin/agent
ENTRYPOINT ["/usr/local/bin/agent"]

FROM debian:bookworm-slim AS server
RUN apt-get update -qq && apt-get install -y -qq ca-certificates >/dev/null && update-ca-certificates
COPY --from=builder /out/server /usr/local/bin/server
ENTRYPOINT ["/usr/local/bin/server"]