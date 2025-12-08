package models

import "time"

// Snapshot is the unified data structure combining all data sources
type Snapshot struct {
	Timestamp   time.Time   `json:"timestamp"`
	Location    string      `json:"location"`
	Weather     Weather     `json:"weather"`
	Environment Environment `json:"environment"`
	Mobility    Mobility    `json:"mobility"`
	Finance     Finance     `json:"finance"`
	Energy      Energy      `json:"energy"`
	Health      Health      `json:"health"`
	Agriculture Agriculture `json:"agriculture"`
	Disasters   Disasters   `json:"disasters"`
}

// Weather holds meteorological data from OpenMeteo
type Weather struct {
	TemperatureC float64 `json:"temperature_c"`
	Humidity     float64 `json:"humidity"`
	WindSpeedMS  float64 `json:"wind_speed_ms"`
	PrecipMM     float64 `json:"precip_mm"`
	CloudCover   float64 `json:"cloud_cover"`
	Visibility   float64 `json:"visibility_km"`
}

// Environment holds air quality data from OpenAQ
type Environment struct {
	PM25  float64 `json:"pm25"`
	PM10  float64 `json:"pm10"`
	Ozone float64 `json:"ozone"`
	NO2   float64 `json:"no2"`
	SO2   float64 `json:"so2"`
	CO    float64 `json:"co"`
}

// Mobility holds transportation data from HERE, OpenSky, and Movebank
type Mobility struct {
	// Traffic (HERE Maps)
	TrafficSpeedKmH  float64 `json:"traffic_speed_kmh"`
	TrafficJamFactor float64 `json:"traffic_jam_factor"`

	// Aviation (OpenSky)
	FlightCount  int     `json:"flight_count"`
	AvgAltitudeM float64 `json:"avg_altitude_m"`

	// Animal Migration (Movebank)
	ActiveSpecies         int     `json:"active_species"`
	AnimalsTracked        int     `json:"animals_tracked"`
	AvgMigrationPaceKMDay float64 `json:"avg_migration_pace_km_day"`
}

// Finance holds financial data from AlphaVantage, NASDAQ
type Finance struct {
	StockPrice      float64 `json:"stock_price"`
	StockSymbol     string  `json:"stock_symbol"`
	CommodityPrice  float64 `json:"commodity_price"`
	CommoditySymbol string  `json:"commodity_symbol"`
	MarketCap       float64 `json:"market_cap"`
	Volume          int64   `json:"volume"`
	NASDAQIndex     float64 `json:"nasdaq_index"`
	VolumeTraded    int64   `json:"volume_traded"`
}

// Energy holds power grid data from Grid, US Energy Info, Ember
type Energy struct {
	ElectricityPriceUSD    float64 `json:"electricity_price_usd"`
	GenerationMWh          float64 `json:"generation_mwh"`
	RenewablePercent       float64 `json:"renewable_percent"`
	GridLoad               float64 `json:"grid_load"`
	CarbonIntensity        float64 `json:"carbon_intensity_gco2_kwh"`
	GridUtilizationPercent float64 `json:"grid_utilization_percent"`
	NaturalGasPriceMmbtu   float64 `json:"natural_gas_price_mmbtu"`
	CoalPercent            float64 `json:"coal_percent"`
	GasPercent             float64 `json:"gas_percent"`
	NuclearPercent         float64 `json:"nuclear_percent"`
}

// Health holds public health data from CDC FluView
type Health struct {
	FluCases           int     `json:"flu_cases"`
	ILIPercent         float64 `json:"ili_percent"` // Influenza-like illness
	HospitalAdmissions int     `json:"hospital_admissions"`
}

// Agriculture holds crop data from USDA NASS
type Agriculture struct {
	CropYield        float64 `json:"crop_yield"`
	CropType         string  `json:"crop_type"`
	SoilMoisture     float64 `json:"soil_moisture_percent"`
	PrecipForecast   float64 `json:"precip_forecast_mm"`
	ProductionBushels float64 `json:"production_bushels"`
	PricePerBushel   float64 `json:"price_per_bushel"`
	HarvestedAcres   float64 `json:"harvested_acres"`
}

// Disasters holds emergency data from FEMA
type Disasters struct {
	ActiveDisasters  int    `json:"active_disasters"`
	DisasterType     string `json:"disaster_type"`
	Severity         int    `json:"severity"` // 1-5 scale
	AffectedCounties int    `json:"affected_counties"`
}
