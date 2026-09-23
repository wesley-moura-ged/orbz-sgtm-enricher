#!/bin/sh
set -eu

database_directory="${GEOIP_DATABASE_PATH:-/data}"
frequency_seconds="${GEOIPUPDATE_FREQUENCY_SECONDS:-259200}"

# Dokploy can pass these credentials through its protected Environment editor.
# The *_FILE variants keep compatibility with existing Docker Swarm Secrets.
read_credential() {
  value="$1"
  file_path="$2"

  if [ -n "$value" ]; then
    printf '%s' "$value"
    return
  fi

  tr -d '\r\n' < "$file_path"
}

account_id="$(read_credential "${MAXMIND_ACCOUNT_ID:-}" "${MAXMIND_ACCOUNT_ID_FILE:-/run/secrets/maxmind_account_id}")"
license_key="$(read_credential "${MAXMIND_LICENSE_KEY:-}" "${MAXMIND_LICENSE_KEY_FILE:-/run/secrets/maxmind_license_key}")"

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
