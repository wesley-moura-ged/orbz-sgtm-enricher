package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wesley-moura-ged/orbz-sgtm-enricher/internal/enricher"
)

func main() {
	config, err := enricher.LoadConfig()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	geoResolver, err := enricher.NewGeoResolver(config.GeoIPDatabasePath)
	if err != nil {
		log.Printf("GeoLite2 database unavailable; GEO headers will remain empty: %v", err)
	}
	if geoResolver != nil {
		defer geoResolver.Close()
	}

	service := enricher.NewService(config, geoResolver)
	server := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           service.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Printf("orbz-enricher listening on %s", config.ListenAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
