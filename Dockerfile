FROM node:24-alpine AS web
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci
COPY apps/web/ ./
RUN npm run build

FROM golang:1.27-alpine AS api
ARG VERSION=development
WORKDIR /src/apps/api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN rm -rf internal/web/dist
COPY --from=web /src/apps/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=${VERSION}" -o /out/zigzag ./cmd/api
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api /out/zigzag /app/zigzag
COPY --from=api --chown=65532:65532 /out/data /data
ENV DATA_DIR=/data
EXPOSE 8091
CMD ["/app/zigzag"]
