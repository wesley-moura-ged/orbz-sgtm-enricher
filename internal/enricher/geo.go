package enricher

import (
	"net/netip"

	geoip2 "github.com/oschwald/geoip2-golang/v2"
)

type Geo struct {
	Country    string
	Region     string
	City       string
	PostalCode string
}

type GeoResolver struct {
	database *geoip2.Reader
}

func NewGeoResolver(path string) (*GeoResolver, error) {
	database, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &GeoResolver{database: database}, nil
}

func (r *GeoResolver) Close() error {
	return r.database.Close()
}

func (r *GeoResolver) Lookup(ip netip.Addr) Geo {
	if r == nil || r.database == nil || !ip.IsValid() {
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
