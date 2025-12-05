package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/clients"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Load .env file if it exists

	openaqKey := os.Getenv("OPENAQ_API_KEY")
	alphaKey := os.Getenv("ALPHAVANTAGE_API_KEY")

	openaq := clients.NewOpenAQClient(openaqKey)
	alpha := clients.NewAlphaVantageClient(alphaKey)
	meteo := clients.NewOpenMeteoClient()
	bikes := clients.NewCityBikesClient()

	if openaqKey == "" {
		log.Printf("skipping OpenAQ: set OPENAQ_API_KEY to enable call")
	} else {
		// 1. USE COORDINATES INSTEAD OF CITY
		// Los Angeles Coordinates: Lat 34.0549, Lon -118.2426
		// Radius: 10000 meters (10km)
		locations, err := openaq.GetLocationsByCoordinates(34.0549, -118.2426, 10000, 10)
		if err != nil {
			log.Printf("OpenAQ error: %v", err)
			return
		}

		if len(locations.Results) == 0 {
			log.Printf("OpenAQ: No locations found at these coordinates.")
		} else {
			var bestLoc *clients.OpenAQLocation

			// 2. Loop to find an ACTIVE location
			// We check if the last update was within the last 24 hours
			for _, loc := range locations.Results {
				if loc.DatetimeLast == nil {
					continue
				}
				
				// Parse the UTC time string
				lastUpdate, err := time.Parse(time.RFC3339, loc.DatetimeLast.UTC)
				if err != nil {
					continue
				}

				// Check if data is fresh (e.g., less than 24 hours old)
				if time.Since(lastUpdate) < 24*time.Hour {
					// Found a live one!
					bestLoc = &loc
					break
				}
			}

			if bestLoc == nil {
				log.Printf("No active sensors found nearby (checked %d candidates)", len(locations.Results))
			} else {
				log.Printf("Found ACTIVE location: %s (Last updated: %s)", bestLoc.Name, bestLoc.DatetimeLast.Local)

				sensors, err := openaq.GetSensorsByLocationID(bestLoc.ID)
				if err != nil {
					log.Printf("Error fetching sensors: %v", err)
				} else {
					log.Printf("Measurements for %s:", bestLoc.Name)
					
					for _, s := range sensors.Results {
						// Skip sensors that have no recent data
						if s.Latest.Datetime.Local == "" {
							continue
						}

						// Now you have access to the Units directly!
						// s.Parameter.DisplayName handles "PM2.5", "Ozone", etc.
						// s.Parameter.Units handles "µg/m³", "ppm", etc.
						
						name := s.Parameter.DisplayName
						if name == "" { name = s.Parameter.Name } // Fallback

						log.Printf("  - %s: %.2f %s (at %s)", 
							name, 
							s.Latest.Value, 
							s.Parameter.Units, 
							s.Latest.Datetime.Local,
						)
					}
				}
			}
		}
	}

	if alphaKey == "" {
		log.Printf("skipping AlphaVantage: set ALPHAVANTAGE_API_KEY to enable call")
	} else if quote, err := alpha.GetGlobalQuote("IBM"); err != nil {
		log.Printf("AlphaVantage error: %v", err)
	} else {
		log.Printf("AlphaVantage %s price %s (open %s, high %s, low %s)", quote.Quote.Symbol, quote.Quote.Price, quote.Quote.Open, quote.Quote.High, quote.Quote.Low)
	}

	if weather, err := meteo.GetCurrentWeather(40.7128, -74.0060); err != nil {
		log.Printf("OpenMeteo error: %v", err)
	} else {
		log.Printf("OpenMeteo NYC temp %.1f C wind %.1f m/s humidity %.0f%%", weather.Current.Temperature2m, weather.Current.WindSpeed10m, weather.Current.RelativeHumidity)
	}

	if networks, err := bikes.ListNetworks(); err != nil {
		log.Printf("CityBikes error: %v", err)
	} else {
		limit := 3
		if len(networks.Networks) < limit {
			limit = len(networks.Networks)
		}
		for i := 0; i < limit; i++ {
			n := networks.Networks[i]
			log.Printf("CityBikes network: %s (%s, %s)", n.Name, n.Location.City, n.Location.Country)
		}
	}

	fmt.Println("EdgeSight Ingest Service demo calls complete")
}
