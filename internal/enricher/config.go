package enricher

import (
	"fmt"
	"os"
	"strings"
)

const defaultSecretPath = "/run/secrets/orbz_user_id_hmac"

type Config struct {
	ListenAddress      string
	GeoIPDatabasePath string
	UserIDSecret      []byte
}

func LoadConfig() (Config, error) {
	// USER_ID_SECRET allows deployment platforms such as Dokploy to provide the
	// HMAC through their protected Environment editor. USER_ID_SECRET_FILE keeps
	// compatibility with existing Docker Swarm deployments that use Secrets.
	secret := []byte(strings.TrimSpace(os.Getenv("USER_ID_SECRET")))
	if len(secret) == 0 {
		secretPath := envOrDefault("USER_ID_SECRET_FILE", defaultSecretPath)
		fromFile, err := os.ReadFile(secretPath)
		if err != nil {
			return Config{}, fmt.Errorf("read user ID secret: %w", err)
		}
		secret = []byte(strings.TrimSpace(string(fromFile)))
	}
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("user ID secret must contain at least 32 bytes")
	}

	return Config{
		ListenAddress:      envOrDefault("LISTEN_ADDRESS", ":8080"),
		GeoIPDatabasePath: envOrDefault("GEOIP_DATABASE_PATH", "/data/GeoLite2-City.mmdb"),
		UserIDSecret:      secret,
	}, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
