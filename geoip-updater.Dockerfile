FROM debian:bookworm-slim

# The Alpine repository used by the previous image does not publish the
# geoipupdate package. Debian Bookworm provides the official client and keeps
# this wrapper focused on reading Docker Secrets at runtime.
RUN apt-get update \
    && apt-get install --no-install-recommends --yes geoipupdate \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/update-geoip.sh /usr/local/bin/update-geoip
RUN chmod 0755 /usr/local/bin/update-geoip

ENTRYPOINT ["/usr/local/bin/update-geoip"]
