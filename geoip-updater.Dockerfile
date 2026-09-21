FROM alpine:3.21

RUN apk add --no-cache geoipupdate

COPY scripts/update-geoip.sh /usr/local/bin/update-geoip
RUN chmod 0755 /usr/local/bin/update-geoip

ENTRYPOINT ["/usr/local/bin/update-geoip"]
