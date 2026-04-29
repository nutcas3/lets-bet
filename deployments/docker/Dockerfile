FROM golang:1.26-alpine AS builder

ARG SERVICE=gateway

WORKDIR /src

RUN apk add --no-cache git ca-certificates curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Download MaxMind GeoLite2 database (optional - requires license key)
# Skip download for now to avoid build failures
# To enable: Get a free license key from https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
# RUN mkdir -p /src/data && \
#     curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=YOUR_LICENSE_KEY&suffix=tar.gz" | \
#     tar xz -C /src/data --strip-components=1 --wildcards "*.mmdb"

RUN CGO_ENABLED=0 GOOS=linux GOFLAGS="-trimpath" \
    go build -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /src/migrations /app/migrations

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/app"]
