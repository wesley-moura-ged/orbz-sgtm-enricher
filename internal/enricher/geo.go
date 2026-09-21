package enricher

import (
	"net/netip"
	"os"
	"sync"
	"time"

	geoip2 "github.com/oschwald/geoip2-golang/v2"
)

type Geo struct {
	Country    string
	Region     string
	City       string
	PostalCode string
}

type GeoResolver struct {
	path     string
	mu       sync.Mutex
	database *geoip2.Reader
	checked  time.Time
	modified time.Time
}

func NewGeoResolver(path string) (*GeoResolver, error) {
	resolver := &GeoResolver{path: path}
	return resolver, resolver.load()
}

func (r *GeoResolver) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.database == nil {
		return nil
	}
	return r.database.Close()
}

func (r *GeoResolver) Lookup(ip netip.Addr) Geo {
	if r == nil || !ip.IsValid() {
		return Geo{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.shouldRefresh() {
		_ = r.load()
	}
	if r.database == nil {
		return Geo{}
	}

	record, err := r.database.City(ip)
	if err != nil || !record.HasData() {
		return Geo{}
	}

	geo := Geo{
		Country:    record.Country.ISOCode,
		City:       record.City.Names.English,
		PostalCode: record.Postal.Code,
	}
	if len(record.Subdivisions) > 0 {
		geo.Region = record.Subdivisions[0].ISOCode
	}
	return geo
}

// load opens the database when it becomes available. It is intentionally safe
// to call after startup: the updater may finish its first download only after
// this service is already accepting ForwardAuth requests. It also reopens the
// database when the updater writes a newer GeoLite2 file.
func (r *GeoResolver) load() error {
	r.checked = time.Now()
	info, err := os.Stat(r.path)
	if err != nil {
		return err
	}
	if r.database != nil && info.ModTime().Equal(r.modified) {
		return nil
	}
	database, err := geoip2.Open(r.path)
	if err != nil {
		return err
	}
	if r.database != nil {
		_ = r.database.Close()
	}
	r.database = database
	r.modified = info.ModTime()
	return nil
}

func (r *GeoResolver) shouldRefresh() bool {
	if r.database == nil {
		return time.Since(r.checked) >= 5*time.Second
	}
	return time.Since(r.checked) >= time.Hour
}
