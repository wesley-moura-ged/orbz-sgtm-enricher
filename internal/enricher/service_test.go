package enricher

import (
	"net/netip"
	"net/http/httptest"
	"testing"
)

func TestClientIPUsesFirstForwardedAddress(t *testing.T) {
	req := httptest.NewRequest("GET", "http://enricher/auth", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")

	ip, ok := clientIP(req)
	if !ok || ip.String() != "203.0.113.9" {
		t.Fatalf("unexpected client IP: %v, valid: %v", ip, ok)
	}
}

func TestUserIDIsStableAndDoesNotExposeInput(t *testing.T) {
	first := userID([]byte("12345678901234567890123456789012"), mustIP(t, "8.8.8.8"), "Example UA", "sgtm.orbztech.com.br")
	second := userID([]byte("12345678901234567890123456789012"), mustIP(t, "8.8.8.8"), "Example UA", "sgtm.orbztech.com.br")

	if first != second || len(first) != 64 {
		t.Fatalf("unexpected user ID: %q", first)
	}
	if first == "8.8.8.8" {
		t.Fatal("user ID exposed the IP address")
	}
}

func TestAuthorizeReturnsNoHeaderForPrivateIP(t *testing.T) {
	service := NewService(Config{UserIDSecret: []byte("12345678901234567890123456789012")}, nil)
	req := httptest.NewRequest("GET", "http://enricher/auth", nil)
	req.Header.Set("X-Forwarded-For", "10.0.1.6")
	response := httptest.NewRecorder()

	service.Handler().ServeHTTP(response, req)
	if response.Code != 204 || response.Header().Get(HeaderUserID) != "" {
		t.Fatalf("unexpected response: code=%d user_id=%q", response.Code, response.Header().Get(HeaderUserID))
	}
}

func mustIP(t *testing.T, raw string) netip.Addr {
	t.Helper()
	ip, err := netip.ParseAddr(raw)
	if err != nil {
		t.Fatal(err)
	}
	return ip
}
