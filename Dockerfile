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
RUN go build -o dbsight .

FROM alpine:3.19
COPY --from=go-builder /app/dbsight /usr/local/bin/
ENTRYPOINT ["dbsight"]
CMD ["serve"]
