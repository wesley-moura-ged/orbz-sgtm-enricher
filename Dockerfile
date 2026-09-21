# Build a small static binary.
FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/orbz-enricher ./cmd/orbz-enricher

# The runtime only receives HTTP from Traefik and reads the GeoLite2 file.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/orbz-enricher /app/orbz-enricher

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/orbz-enricher"]
