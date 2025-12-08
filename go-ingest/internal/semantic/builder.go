package semantic

import (
	"fmt"
	"strings"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/models"
)

// GenerateSummary creates a natural language description of a snapshot
func GenerateSummary(snap models.Snapshot) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Location: %s at %s", snap.Location, snap.Timestamp.Format("Jan 02, 2006 3:04 PM MST")))

	// Weather
	if snap.Weather.TemperatureC != 0 || snap.Weather.Humidity != 0 {
		weather := fmt.Sprintf("Weather: %.1f°C, %.0f%% humidity, wind %.1f m/s",
			snap.Weather.TemperatureC, snap.Weather.Humidity, snap.Weather.WindSpeedMS)
		if snap.Weather.PrecipMM > 0 {
			weather += fmt.Sprintf(", %.1fmm precipitation", snap.Weather.PrecipMM)
		}
		parts = append(parts, weather)
	}

	// Air Quality
	if snap.Environment.PM25 > 0 || snap.Environment.PM10 > 0 {
		aq := fmt.Sprintf("Air Quality: PM2.5 %.1f µg/m³ (%s), PM10 %.1f µg/m³",
			snap.Environment.PM25, interpretAQI(snap.Environment.PM25), snap.Environment.PM10)
		if snap.Environment.Ozone > 0 {
			aq += fmt.Sprintf(", O₃ %.2f ppm", snap.Environment.Ozone)
		}
		parts = append(parts, aq)
	}

	// Mobility
	if snap.Mobility.TrafficSpeedKmH > 0 {
		parts = append(parts, fmt.Sprintf("Traffic: avg speed %.1f km/h, jam factor %.2f",
			snap.Mobility.TrafficSpeedKmH, snap.Mobility.TrafficJamFactor))
	}
	if snap.Mobility.FlightCount > 0 {
		parts = append(parts, fmt.Sprintf("Aviation: %d flights overhead, avg altitude %.0fm",
			snap.Mobility.FlightCount, snap.Mobility.AvgAltitudeM))
	}
	if snap.Mobility.ActiveSpecies > 0 || snap.Mobility.AnimalsTracked > 0 {
		parts = append(parts, fmt.Sprintf("Wildlife: %d species, %d animals tracked, %.1f km/day pace",
			snap.Mobility.ActiveSpecies, snap.Mobility.AnimalsTracked, snap.Mobility.AvgMigrationPaceKMDay))
	}

	// Finance
	if snap.Finance.StockPrice > 0 {
		parts = append(parts, fmt.Sprintf("Equity: %s at $%.2f",
			snap.Finance.StockSymbol, snap.Finance.StockPrice))
	}
	if snap.Finance.NASDAQIndex > 0 {
		parts = append(parts, fmt.Sprintf("NASDAQ: %.2f (vol %d)", snap.Finance.NASDAQIndex, snap.Finance.VolumeTraded))
	}
	if snap.Finance.CommodityPrice > 0 {
		parts = append(parts, fmt.Sprintf("Commodity: %s at $%.2f",
			snap.Finance.CommoditySymbol, snap.Finance.CommodityPrice))
	}

	// Energy
	if snap.Energy.ElectricityPriceUSD > 0 || snap.Energy.GenerationMWh > 0 || snap.Energy.RenewablePercent > 0 {
		parts = append(parts, fmt.Sprintf("Energy: $%.4f/kWh, %.0f MWh gen, %.1f%% renewable, CI %.0f gCO2/kWh",
			snap.Energy.ElectricityPriceUSD, snap.Energy.GenerationMWh, snap.Energy.RenewablePercent, snap.Energy.CarbonIntensity))
	}

	// Health
	if snap.Health.FluCases > 0 || snap.Health.ILIPercent > 0 {
		parts = append(parts, fmt.Sprintf("Health: %d cases, %.1f%% ILI/RSV",
			snap.Health.FluCases, snap.Health.ILIPercent))
	}

	// Agriculture
	if snap.Agriculture.CropYield > 0 {
		parts = append(parts, fmt.Sprintf("Agriculture: %s yield %.1f, soil moisture %.1f%%",
			snap.Agriculture.CropType, snap.Agriculture.CropYield, snap.Agriculture.SoilMoisture))
	}

	// Disasters
	if snap.Disasters.ActiveDisasters > 0 {
		parts = append(parts, fmt.Sprintf("⚠️ Disasters: %d active (%s, severity %d), %d counties affected",
			snap.Disasters.ActiveDisasters, snap.Disasters.DisasterType, snap.Disasters.Severity, snap.Disasters.AffectedCounties))
	}

	return strings.Join(parts, ". ")
}

// interpretAQI converts PM2.5 µg/m³ to qualitative category
func interpretAQI(pm25 float64) string {
	if pm25 <= 12.0 {
		return "Good"
	} else if pm25 <= 35.4 {
		return "Moderate"
	} else if pm25 <= 55.4 {
		return "Unhealthy for Sensitive"
	} else if pm25 <= 150.4 {
		return "Unhealthy"
	} else if pm25 <= 250.4 {
		return "Very Unhealthy"
	}
	return "Hazardous"
}
