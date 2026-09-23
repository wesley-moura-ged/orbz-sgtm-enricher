package enricher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadConfigAcceptsEnvironmentSecret(t *testing.T) {
	t.Setenv("USER_ID_SECRET", "12345678901234567890123456789012")
	t.Setenv("USER_ID_SECRET_FILE", "/this-file-must-not-be-read")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if string(config.UserIDSecret) != "12345678901234567890123456789012" {
		t.Fatalf("unexpected user ID secret: %q", config.UserIDSecret)
	}
}

func TestForwardAuthEnrichesRequest(t *testing.T) {
	secret := []byte("test-secret-that-has-more-than-thirty-two-characters")
	service := NewService(Config{
		GeoIPDatabasePath: "",
		UserIDSecret:      secret,
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.test/?sck=abc", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	recorder := httptest.NewRecorder()

	service.ForwardAuth(recorder, req)

	response := recorder.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	expectedMAC := hmac.New(sha256.New, secret)
	_, _ = expectedMAC.Write([]byte("abc"))
	expectedID := hex.EncodeToString(expectedMAC.Sum(nil))

	if got := response.Header.Get("X-Orbz-User-Id"); got != expectedID {
		t.Fatalf("X-Orbz-User-Id = %q, want %q", got, expectedID)
	}

	if got := response.Header.Get("X-Orbz-Geo-Status"); got != "unavailable" {
		t.Fatalf("X-Orbz-Geo-Status = %q, want unavailable", got)
	}
}

func TestForwardAuthWithoutSCKDoesNotSetUserID(t *testing.T) {
	service := NewService(Config{
		UserIDSecret: []byte("test-secret-that-has-more-than-thirty-two-characters"),
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	recorder := httptest.NewRecorder()

	service.ForwardAuth(recorder, req)

	if got := recorder.Result().Header.Get("X-Orbz-User-Id"); got != "" {
		t.Fatalf("X-Orbz-User-Id = %q, want empty", got)
	}
}
