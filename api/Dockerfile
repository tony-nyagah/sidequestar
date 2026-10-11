FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The SQLite driver is pure Go, so the binary is static.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sidequestar-api ./cmd/server

FROM alpine:3
RUN adduser -D -u 10001 app && mkdir -p /data && chown app /data
COPY --from=build /out/sidequestar-api /usr/local/bin/sidequestar-api
USER app
ENV DATA_DIR=/data API_HOST=0.0.0.0 PORT=8080
EXPOSE 8080
ENTRYPOINT ["sidequestar-api"]
