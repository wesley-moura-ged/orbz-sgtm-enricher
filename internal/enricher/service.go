package enricher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/netip"
	"strings"
)

const (
	HeaderUserID = "X-Orbz-User-Id"
	HeaderCountry = "X-GEO-Country"
	HeaderRegion = "X-GEO-Region"
	HeaderCity = "X-GEO-City"
	HeaderPostal = "X-GEO-PostalCode"
)

type Service struct {
	config Config
	geo    *GeoResolver
}

func NewService(config Config, geo *GeoResolver) *Service {
	return &Service{config: config, geo: geo}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("/auth", s.authorize)
	return mux
}

func (s *Service) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
}

// authorize is the endpoint consumed by Traefik ForwardAuth. It never proxies
// the request; a successful 204 instructs Traefik to continue the original one.
func (s *Service) authorize(w http.ResponseWriter, r *http.Request) {
	ip, ok := clientIP(r)
	if !ok || !ip.IsGlobalUnicast() {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set(HeaderUserID, userID(s.config.UserIDSecret, ip, r.UserAgent(), requestHost(r)))
	geo := s.geo.Lookup(ip)
	setHeaderIfPresent(w, HeaderCountry, geo.Country)
	setHeaderIfPresent(w, HeaderRegion, geo.Region)
	setHeaderIfPresent(w, HeaderCity, geo.City)
	setHeaderIfPresent(w, HeaderPostal, geo.PostalCode)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

func clientIP(r *http.Request) (netip.Addr, bool) {
	// Traefik must sanitize and trust forwarded headers at the entry point.
	// The first X-Forwarded-For value is the original visitor when this contract holds.
	for _, candidate := range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		if ip, err := netip.ParseAddr(strings.TrimSpace(candidate)); err == nil {
			return ip.Unmap(), true
		}
	}
	if ip, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("X-Real-Ip"))); err == nil {
		return ip.Unmap(), true
	}
	return netip.Addr{}, false
}

func requestHost(r *http.Request) string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	return strings.ToLower(host)
}

func userID(secret []byte, ip netip.Addr, userAgent, host string) string {
	message := strings.Join([]string{ip.String(), strings.TrimSpace(userAgent), host}, "\n")
	hash := hmac.New(sha256.New, secret)
	_, _ = hash.Write([]byte(message))
	return hex.EncodeToString(hash.Sum(nil))
}

func setHeaderIfPresent(w http.ResponseWriter, name, value string) {
	if strings.TrimSpace(value) != "" {
		w.Header().Set(name, value)
	}
}
