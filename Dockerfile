# syntax=docker/dockerfile:1.7

FROM node:22-bookworm-slim AS frontend-builder

WORKDIR /src/web

COPY web/package.json web/pnpm-lock.yaml ./

RUN corepack enable && pnpm install --frozen-lockfile

COPY web/ ./

RUN pnpm build


FROM golang:1.25-bookworm AS backend-builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY --from=frontend-builder /src/web/dist ./web/dist

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG APP_VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
	-trimpath \
	-tags embed_frontend \
	-ldflags="-s -w -X main.appVersion=${APP_VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
	-o /out/easydrop .


FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=backend-builder /out/easydrop /app/easydrop

EXPOSE 8080

VOLUME ["/app/data"]

ENTRYPOINT ["/app/easydrop"]
CMD ["--auto-generate-jwt"]
