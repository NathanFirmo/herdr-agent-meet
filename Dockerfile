# The web bundle and the Go binary are built on the host platform and the
# binary is cross-compiled, so multi-platform builds only emulate the last stage.
FROM --platform=$BUILDPLATFORM node:24-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS api
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /agent-meet ./cmd/agent-meet

# Debian (glibc) so CLIs mounted from the host, like a summarizer, can run.
FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=api /agent-meet /usr/local/bin/agent-meet
COPY --from=web /web/dist /app/web
ENV STATIC_DIR=/app/web \
    LISTEN_ADDR=:8765 \
    HERDR_SOCKET_PATH=/run/herdr/herdr.sock \
    HOME=/tmp
EXPOSE 8765
ENTRYPOINT ["agent-meet"]
