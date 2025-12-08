package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/canonicalizer"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/clients"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/embeddings"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/semantic"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Load .env file if it exists

	openaqKey := os.Getenv("OPENAQ_API_KEY")
	alphaKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	femaJSONPath := os.Getenv("FEMA_JSON_PATH")
	femaState := os.Getenv("FEMA_STATE_CODE")
	if femaState == "" {
		femaState = "CA"
	}

	femaLookbackDays := 180
	if envDays := os.Getenv("FEMA_LOOKBACK_DAYS"); envDays != "" {
		if days, err := strconv.Atoi(envDays); err == nil && days > 0 {
			femaLookbackDays = days
		}
	}

	// Initialize database
	db, err := store.NewSQLiteStore("edgesight.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	openaq := clients.NewOpenAQClient(openaqKey)
	alpha := clients.NewAlphaVantageClient(alphaKey)
	meteo := clients.NewOpenMeteoClient()
	fema := clients.NewFEMAClient(femaJSONPath)
	cdc := clients.NewCDCFluViewClient()
	nrevssCSV := os.Getenv("NREVSS_CSV_PATH")
	movebankUser := os.Getenv("MOVEBANK_USERNAME")
	movebankPass := os.Getenv("MOVEBANK_PASSWORD")
	movebank := clients.NewMovebankClient(movebankUser, movebankPass)
	stooq := clients.NewStooqClient()
	fredKey := os.Getenv("FRED_API_KEY")
	var fred *clients.FREDClient
	if fredKey != "" {
		fred = clients.NewFREDClient(fredKey)
	}
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "tcp://localhost:1883"
	}
	mqttCli := clients.NewMQTTSensorClient(mqttBroker)

	embedEndpoint := os.Getenv("EMBEDDING_ENDPOINT")
	if embedEndpoint == "" {
		embedEndpoint = "http://localhost:9000"
	}
	var embedCli *embeddings.Client
	if embedEndpoint != "" {
		embedCli = embeddings.NewClient(embedEndpoint)
	}

	ember := clients.NewEmberClient()
	grid := clients.NewGridClient("CAISO") // California ISO

	eiaKey := os.Getenv("EIA_API_KEY")
	var eia *clients.EIAClient
	if eiaKey != "" {
		eia = clients.NewEIAClient(eiaKey)
	}

	nassKey := os.Getenv("NASS_API_KEY")
	var nass *clients.NASSClient
	if nassKey != "" {
		nass = clients.NewNASSClient(nassKey)
	}

	// Variables to collect for snapshot
	var meteoData *clients.CurrentWeatherResponse
	var sensorsData *clients.SensorsResponse
	var stockPrice float64 = 0
	var nasdaqData *clients.NASDAQMarketSummary
	var emberData *clients.EmberElectricitySummary
	var gridData *clients.GridStatus
	var eiaData *clients.EIAEnergySummary
	var nassData *clients.NASSCropSummary
	var disastersData *clients.FEMASummary
	var fluData *clients.CDCFluSummary
	var movementData *clients.MovementSummary
	location := "Los Angeles"

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
					sensorsData = sensors
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
						if name == "" {
							name = s.Parameter.Name
						} // Fallback

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
		priceFloat, _ := strconv.ParseFloat(quote.Quote.Price, 64)
		stockPrice = priceFloat
		log.Printf("AlphaVantage %s price %s (open %s, high %s, low %s)", quote.Quote.Symbol, quote.Quote.Price, quote.Quote.Open, quote.Quote.High, quote.Quote.Low)
	}

	if weather, err := meteo.GetCurrentWeather(40.7128, -74.0060); err != nil {
		log.Printf("OpenMeteo error: %v", err)
	} else {
		meteoData = weather
		log.Printf("OpenMeteo NYC temp %.1f C wind %.1f m/s humidity %.0f%%", weather.Current.Temperature2m, weather.Current.WindSpeed10m, weather.Current.RelativeHumidity)
	}

	if summary, err := fema.GetStateSummary(femaState, femaLookbackDays); err != nil {
		log.Printf("FEMA error: %v", err)
	} else {
		disastersData = summary
		log.Printf("FEMA %s: %d active (%s), %d counties", femaState, summary.ActiveDisasters, summary.TopIncidentType, summary.AffectedCounties)
	}

	if nrevssCSV != "" {
		if fluSummary, err := cdc.GetNREVSSSummaryFromCSV(nrevssCSV); err != nil {
			log.Printf("NREVSS CSV error: %v", err)
		} else {
			fluData = fluSummary
			log.Printf("NREVSS RSV: %.2f%% positive, %d detections, %d tests (week ending %s)", fluSummary.UnweightedILI, fluSummary.FluCases, fluSummary.HospitalAdmissions, fluSummary.WeekEndDate.Format("2006-01-02"))
		}
	} else if fluSummary, err := cdc.GetNationalILIData(); err != nil {
		log.Printf("CDC FluView error: %v", err)
	} else {
		fluData = fluSummary
		log.Printf("CDC ILI: %.2f%% unweighted ILI, %d cases, %d hospitalizations", fluSummary.UnweightedILI, fluSummary.FluCases, fluSummary.HospitalAdmissions)
	}

	// MQTT simulated sensors (non-fatal if broker unavailable)
	var mqttData *clients.MQTTSensorReading
	if mqttCli != nil {
		if m, err := mqttCli.FetchReadings(); err != nil {
			log.Printf("MQTT error: %v", err)
		} else {
			mqttData = m
			log.Printf("MQTT sensors: temp %.1fC, humidity %.0f%%, PM2.5 %.1f, power %.0f",
				m.Temperature, m.Humidity, m.PM25, m.Power)
		}
	}

	if movement, err := movebank.GetGlobalMovementTrends(); err != nil {
		log.Printf("Movebank error: %v", err)
	} else {
		movementData = movement
		log.Printf("Movebank: %d species, %d animals tracked, %.1f km/day avg migration pace", movement.ActiveSpecies, movement.TotalAnimalsTracked, movement.AvgMigrationPace)
	}

	// Market index: prefer FRED (official) if key present; otherwise Stooq
	if fred != nil {
		if market, err := fred.GetNasdaqComposite(); err != nil {
			log.Printf("FRED NASDAQ error: %v", err)
			if stooqMarket, err2 := stooq.GetNasdaqComposite(); err2 != nil {
				log.Printf("Stooq NASDAQ error: %v", err2)
			} else {
				nasdaqData = stooqMarket
				log.Printf("Stooq NASDAQ: %.2f, Volume: %d", stooqMarket.IndexValue, stooqMarket.VolumeTraded)
			}
		} else {
			nasdaqData = market
			log.Printf("FRED NASDAQ: %.2f", market.IndexValue)
		}
	} else {
		if stooqMarket, err := stooq.GetNasdaqComposite(); err != nil {
			log.Printf("Stooq NASDAQ error: %v", err)
		} else {
			nasdaqData = stooqMarket
			log.Printf("Stooq NASDAQ: %.2f, Volume: %d", stooqMarket.IndexValue, stooqMarket.VolumeTraded)
		}
	}

	if summary, err := ember.GetGlobalAverage(); err != nil {
		log.Printf("Ember error: %v", err)
	} else {
		emberData = summary
		log.Printf("Ember Global: %.1f gCO2/kWh carbon intensity, %.1f%% renewable", summary.CarbonIntensityGCO2KWh, summary.RenewablePercent)
	}

	if status, err := grid.GetGridStatus(); err != nil {
		log.Printf("Grid error: %v", err)
	} else {
		gridData = status
		log.Printf("Grid Status: %.0f MW load (%.1f%% utilization), %s", status.LoadMW, status.UtilizationPercent, status.Status)
	}

	if eia != nil {
		if energySummary, err := eia.GetEnergySummary(); err != nil {
			log.Printf("EIA error: %v", err)
		} else {
			eiaData = energySummary
			log.Printf("EIA: %.0f MWh generation, $%.2f/MMBtu natural gas", energySummary.ElectricityGenerationMWh, energySummary.NaturalGasPriceMmbtu)
		}
	} else {
		log.Printf("skipping EIA: set EIA_API_KEY to enable call")
	}

	if nass != nil {
		if cropSummary, err := nass.GetNationalCropSummary("CORN"); err != nil {
			log.Printf("NASS error: %v", err)
		} else {
			nassData = cropSummary
			log.Printf("NASS %s: %.0f bushels, %.1f bu/acre yield, $%.2f/bu", cropSummary.CropType, cropSummary.ProductionBushels, cropSummary.YieldPerAcre, cropSummary.PricePerBushel)
		}
	} else {
		log.Printf("skipping NASS: set NASS_API_KEY to enable call")
	}

	// Build unified snapshot from all sources
	snap := canonicalizer.BuildSnapshot(location, meteoData, sensorsData, mqttData, stockPrice, nasdaqData, emberData, gridData, eiaData, nassData, disastersData, fluData, movementData)

	// Persist to database
	if err := db.InsertSnapshot(snap); err != nil {
		log.Printf("Error inserting snapshot: %v", err)
	} else {
		log.Printf("Snapshot stored in database for %s at %s", snap.Location, snap.Timestamp.Format(time.RFC3339))
	}

	// Generate and store embedding (best-effort)
	if embedCli != nil {
		summary := semantic.GenerateSummary(snap)
		if vec, err := embedCli.Embed(summary); err != nil {
			log.Printf("Embedding error: %v", err)
		} else {
			e := store.SnapshotEmbedding{
				SnapshotTS: snap.Timestamp.Format(time.RFC3339),
				Location:   snap.Location,
				Summary:    summary,
				Embedding:  vec,
				CreatedAt:  time.Now().UTC(),
			}
			if err := db.InsertEmbedding(e); err != nil {
				log.Printf("Insert embedding error: %v", err)
			}
		}
	}

	fmt.Println("EdgeSight Ingest Service demo calls complete")
}
