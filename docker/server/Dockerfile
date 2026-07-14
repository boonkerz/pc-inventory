# Multi-Stage-Build des Inventory-Servers (Frontend + Go) zu einem schlanken Image.
# Build-Kontext ist das Repo-Wurzelverzeichnis (siehe compose.yml: context: ..).

# 1) Frontend bauen
FROM node:20-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# 2) Server bauen (CGo-frei -> statisches Binary)
FROM golang:1.26-alpine AS build
ARG VERSION=0.0.0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# frisch gebautes Frontend einbetten (überschreibt den Platzhalter in web/dist)
COPY --from=web /web/dist ./web/dist
# Agent-Binaries für alle Plattformen cross-kompilieren und in den Server einbetten,
# damit Clients den Agent direkt vom Inventory-Server herunterladen können.
# Server UND eingebettete Agents tragen dieselbe Version (für das Auto-Update).
RUN mkdir -p internal/server/agentdist/bin && \
    CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o internal/server/agentdist/bin/agent-linux-amd64       ./cmd/agent && \
    CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w -X main.version=$VERSION" -o internal/server/agentdist/bin/agent-linux-arm64       ./cmd/agent && \
    CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o internal/server/agentdist/bin/agent-darwin-amd64      ./cmd/agent && \
    CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w -X main.version=$VERSION" -o internal/server/agentdist/bin/agent-darwin-arm64      ./cmd/agent && \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o internal/server/agentdist/bin/agent-windows-amd64.exe ./cmd/agent
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$VERSION" -o /server ./cmd/server

# 3) Laufzeit-Image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && mkdir -p /var/lib/roster
COPY --from=build /server /usr/local/bin/server
EXPOSE 8443
ENTRYPOINT ["server", "run"]
