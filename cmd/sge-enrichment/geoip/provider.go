package geoip

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type Location struct {
	Country string
	City    string
	ISO     string
	Lat     float64
	Lon     float64
}

type Provider struct {
	db *geoip2.Reader
}

func NewProvider(path string) (*Provider, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		log.Printf("[GeoIP] Warning: DB not found at %s. Geo enrichment disabled.", path)
		return &Provider{db: nil}, nil // Return nil db but no error to allow start
	}
	return &Provider{db: db}, nil
}

func (p *Provider) Lookup(ipStr string) *Location {
	if p.db == nil {
		return nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}

	record, err := p.db.City(ip)
	if err != nil {
		return nil
	}

	return &Location{
		Country: record.Country.Names["en"],
		City:    record.City.Names["en"],
		ISO:     record.Country.IsoCode,
		Lat:     record.Location.Latitude,
		Lon:     record.Location.Longitude,
	}
}

func (p *Provider) Close() {
	if p.db != nil {
		p.db.Close()
	}
}
