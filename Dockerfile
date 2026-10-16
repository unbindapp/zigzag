# syntax=docker/dockerfile:1
FROM node:24-alpine AS web
ENV PNPM_HOME=/pnpm PATH=/pnpm:$PATH COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable && corepack prepare pnpm@10.34.7 --activate && pnpm config set store-dir /pnpm/store && pnpm config set update-notifier false
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
COPY apps/web/index.html apps/web/vite.config.ts apps/web/tsconfig.json ./
COPY apps/web/public ./public
COPY apps/web/src ./src
RUN pnpm run build

FROM golang:1.27-alpine AS deps
ENV CGO_ENABLED=0 GOCACHE=/gocache GOFLAGS="-trimpath -buildvcs=false"
WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
RUN GOFLAGS="$GOFLAGS -mod=mod" go list -e -f '{{if and (not .Error) (not .DepsErrors) (not .Standard)}}{{.ImportPath}}{{end}}' $(go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}/...{{end}}' all) 2>/dev/null \
    | grep -vE '/cmd/|/example|/testdata' \
    | xargs -n 20 sh -c 'GOFLAGS="$GOFLAGS -mod=mod" go build "$@" >/dev/null 2>&1 || true' sh

FROM deps AS api
ARG VERSION=development
COPY apps/api/ ./
COPY --from=web /src/apps/web/dist ./internal/web/dist
RUN --mount=type=cache,target=/gocache,from=deps,source=/gocache \
    go build -ldflags "-s -w -X main.Version=${VERSION}" -o /out/zigzag ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api /out/zigzag /app/zigzag
EXPOSE 8091
CMD ["/app/zigzag"]
