FROM node:24-alpine AS web-builder
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /build

# Copy workspace manifests first (cache layer)
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY apps/web/package.json ./apps/web/

# Install deps
RUN pnpm install --frozen-lockfile

# Copy source and build
COPY apps/web/ ./apps/web/
RUN pnpm --filter web build

FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
COPY --from=web-builder /build/apps/web/dist ./apps/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dbsight .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -h /app appuser
WORKDIR /app
COPY --from=go-builder /app/dbsight .
USER appuser
EXPOSE 42198
ENTRYPOINT ["./dbsight"]
CMD ["serve"]
