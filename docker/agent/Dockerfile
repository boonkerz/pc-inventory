# Linux-Test-PC: baut den Agent und startet ihn als Vordergrundprozess.
# Build-Kontext ist das Repo-Wurzelverzeichnis (siehe compose.yml).

FROM golang:1.26-alpine AS build
ARG VERSION=0.0.0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$VERSION" -o /agent ./cmd/agent

FROM alpine:3.20
# ca-certificates für TLS; util-linux/dmidecode optional für mehr Hardware-Infos.
RUN apk add --no-cache ca-certificates
COPY --from=build /agent /usr/local/bin/agent
COPY docker/agent/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
