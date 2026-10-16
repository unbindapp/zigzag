# syntax=docker/dockerfile:1
FROM node:24-alpine AS web
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci
COPY apps/web/ ./
RUN npm run build

FROM golang:1.27-alpine AS deps
ENV CGO_ENABLED=0 GOCACHE=/gocache GOFLAGS=-trimpath
WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
# Compiles the direct dependencies into GOCACHE so the layer, keyed on go.mod and go.sum, carries the compiled tree between builds.
RUN GOFLAGS="-trimpath -mod=mod" go list -e -f '{{if and (not .Error) (not .DepsErrors) (not .Standard)}}{{.ImportPath}}{{end}}' $(go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}/...{{end}}' all) 2>/dev/null \
    | xargs -n 20 sh -c 'GOFLAGS="-trimpath -mod=mod" go build "$@" >/dev/null 2>&1 || true' sh

FROM deps AS api
ARG VERSION=development
COPY apps/api/ ./
RUN rm -rf internal/web/dist
COPY --from=web /src/apps/web/dist ./internal/web/dist
RUN --mount=type=cache,target=/gocache,from=deps,source=/gocache \
    go build -ldflags "-s -w -X main.Version=${VERSION}" -o /out/zigzag ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api /out/zigzag /app/zigzag
EXPOSE 8091
CMD ["/app/zigzag"]
