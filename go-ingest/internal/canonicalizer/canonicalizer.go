package canonicalizer

import (
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/clients"
	"github.com/ColonelToad/EdgeSight/go-ingest/internal/models"
)

// BuildSnapshot unifies data from all sources into a single Snapshot.
// Why this structure:
// - OpenMeteo: current weather (temp, humidity, wind)
// - OpenAQ: sensors with latest readings (PM2.5, PM10, Ozone, etc.)
// - AlphaVantage: stock price
// - NASDAQ: market composite index
// - Ember: carbon intensity and generation mix
// - Grid: power grid status and load
// - EIA: US energy generation and prices
// - NASS: USDA crop production and prices
// - FEMA: disaster declarations
// - CDC FluView: influenza surveillance
// - Movebank: animal migration/movement trends
func BuildSnapshot(
	location string,
	meteo *clients.CurrentWeatherResponse,
	sensors *clients.SensorsResponse,
	mqttData *clients.MQTTSensorReading,
	stockPrice float64,
	nasdaq *clients.NASDAQMarketSummary,
	ember *clients.EmberElectricitySummary,
	grid *clients.GridStatus,
	eia *clients.EIAEnergySummary,
	nass *clients.NASSCropSummary,
	disasters *clients.FEMASummary,
	fluSummary *clients.CDCFluSummary,
	movementSummary *clients.MovementSummary,
) models.Snapshot {

	snap := models.Snapshot{
		Timestamp: time.Now().UTC(),
		Location:  location,
	}

	// --- Weather: from OpenMeteo current block ---
	if meteo != nil {
		snap.Weather.TemperatureC = meteo.Current.Temperature2m
		snap.Weather.Humidity = meteo.Current.RelativeHumidity
		snap.Weather.WindSpeedMS = meteo.Current.WindSpeed10m
		// OpenMeteo doesn't provide precip in the "current" block by default,
		// but you could extend it if needed
	}

	// --- Environment: from OpenAQ sensors ---
	// OpenAQ v3 returns a SensorsResponse with each sensor containing:
	// - Parameter (DisplayName, Units, Name)
	// - Latest (Value, Datetime)
	if sensors != nil {
		for _, sensor := range sensors.Results {
			// Skip sensors with no recent data
			if sensor.Latest.Datetime.Local == "" {
				continue
			}

			paramName := normalizeAQParam(sensor.Parameter.Name)
			switch paramName {
			case "pm25":
				snap.Environment.PM25 = sensor.Latest.Value
			case "pm10":
				snap.Environment.PM10 = sensor.Latest.Value
			case "o3":
				snap.Environment.Ozone = sensor.Latest.Value
			}
		}
	}

	// --- Environment: from MQTT simulated sensors (overrides if present) ---
	if mqttData != nil {
		if mqttData.PM25 > 0 {
			snap.Environment.PM25 = mqttData.PM25
		}
		if mqttData.Temperature != 0 {
			snap.Weather.TemperatureC = mqttData.Temperature
		}
		if mqttData.Humidity != 0 {
			snap.Weather.Humidity = mqttData.Humidity
		}
		if mqttData.Power > 0 {
			snap.Energy.GridLoad = mqttData.Power
		}
	}

	// --- Finance ---
	snap.Finance.StockPrice = stockPrice

	// --- Finance: from NASDAQ Data Link ---
	if nasdaq != nil {
		snap.Finance.NASDAQIndex = nasdaq.IndexValue
		snap.Finance.VolumeTraded = nasdaq.VolumeTraded
	}

	// --- Energy: from Ember Climate ---
	if ember != nil {
		snap.Energy.CarbonIntensity = ember.CarbonIntensityGCO2KWh
		snap.Energy.RenewablePercent = ember.RenewablePercent
		snap.Energy.GenerationMWh = ember.GenerationTWh * 1000 // Convert TWh to MWh
		snap.Energy.CoalPercent = ember.CoalPercent
		snap.Energy.GasPercent = ember.GasPercent
		snap.Energy.NuclearPercent = ember.NuclearPercent
	}

	// --- Energy: from Grid monitoring ---
	if grid != nil {
		snap.Energy.GridLoad = grid.LoadMW
		snap.Energy.GridUtilizationPercent = grid.UtilizationPercent
	}

	// --- Energy: from EIA (US Energy Information Administration) ---
	if eia != nil {
		snap.Energy.GenerationMWh = eia.ElectricityGenerationMWh
		snap.Energy.NaturalGasPriceMmbtu = eia.NaturalGasPriceMmbtu
		// EIA can override Ember data if available
		if eia.RenewableGenerationMWh > 0 && eia.ElectricityGenerationMWh > 0 {
			snap.Energy.RenewablePercent = (eia.RenewableGenerationMWh / eia.ElectricityGenerationMWh) * 100
		}
	}

	// --- Agriculture: from USDA NASS ---
	if nass != nil {
		snap.Agriculture.CropType = nass.CropType
		snap.Agriculture.CropYield = nass.YieldPerAcre
		snap.Agriculture.ProductionBushels = nass.ProductionBushels
		snap.Agriculture.PricePerBushel = nass.PricePerBushel
		snap.Agriculture.HarvestedAcres = nass.HarvestedAcres
	}

	// --- Disasters: from FEMA static JSON ---
	if disasters != nil {
		snap.Disasters.ActiveDisasters = disasters.ActiveDisasters
		snap.Disasters.DisasterType = disasters.TopIncidentType
		snap.Disasters.Severity = disasters.Severity
		snap.Disasters.AffectedCounties = disasters.AffectedCounties
	}

	// --- Health: from CDC FluView ---
	if fluSummary != nil {
		snap.Health.FluCases = fluSummary.FluCases
		snap.Health.ILIPercent = fluSummary.UnweightedILI
		snap.Health.HospitalAdmissions = fluSummary.HospitalAdmissions
	}

	// --- Mobility: Animal migration/movement trends from Movebank ---
	if movementSummary != nil {
		snap.Mobility.ActiveSpecies = movementSummary.ActiveSpecies
		snap.Mobility.AnimalsTracked = movementSummary.TotalAnimalsTracked
		snap.Mobility.AvgMigrationPaceKMDay = movementSummary.AvgMigrationPace
	}

	return snap
}

// normalizeAQParam converts various parameter names to canonical forms
func normalizeAQParam(name string) string {
	switch name {
	case "pm25", "pm2.5", "PM2.5":
		return "pm25"
	case "pm10", "PM10":
		return "pm10"
	case "o3", "ozone", "O3":
		return "o3"
	}
	return name
}
