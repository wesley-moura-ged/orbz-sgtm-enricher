#!/bin/sh
set -eu

account_id_file="${MAXMIND_ACCOUNT_ID_FILE:-/run/secrets/maxmind_account_id}"
license_key_file="${MAXMIND_LICENSE_KEY_FILE:-/run/secrets/maxmind_license_key}"
database_directory="${GEOIP_DATABASE_PATH:-/data}"
frequency_seconds="${GEOIPUPDATE_FREQUENCY_SECONDS:-259200}"

account_id="$(tr -d '\r\n' < "$account_id_file")"
license_key="$(tr -d '\r\n' < "$license_key_file")"

if [ -z "$account_id" ] || [ -z "$license_key" ]; then
  echo "MaxMind credentials are empty" >&2
  exit 1
fi

mkdir -p "$database_directory"
config_file="$(mktemp)"
trap 'rm -f "$config_file"' EXIT INT TERM

cat > "$config_file" <<EOF
AccountID $account_id
LicenseKey $license_key
EditionIDs GeoLite2-City
DatabaseDirectory $database_directory
EOF

while true; do
  # The updater does not print the generated configuration or secret values.
  geoipupdate -f "$config_file" || echo "GeoLite2 update failed; keeping the last valid database" >&2
  sleep "$frequency_seconds"
done
