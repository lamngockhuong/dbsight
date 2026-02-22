FROM node:20-alpine AS web-builder
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm run build

FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
COPY --from=web-builder /web/dist ./web/dist
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
