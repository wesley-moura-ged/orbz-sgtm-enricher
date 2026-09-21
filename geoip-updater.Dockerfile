# Compile the official MaxMind client instead of depending on an operating
# system package whose availability varies between distributions.
FROM golang:1.25-bookworm AS geoipupdate-build

RUN go install github.com/maxmind/geoipupdate/v8/cmd/geoipupdate@v8.0.0

FROM debian:bookworm-slim

# The slim base does not include the public CA bundle. The updater uses HTTPS
# to download the GeoLite2 database from MaxMind.
RUN apt-get update \
    && apt-get install --no-install-recommends --yes ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=geoipupdate-build /go/bin/geoipupdate /usr/local/bin/geoipupdate

COPY scripts/update-geoip.sh /usr/local/bin/update-geoip
RUN chmod 0755 /usr/local/bin/update-geoip

ENTRYPOINT ["/usr/local/bin/update-geoip"]
