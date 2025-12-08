package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/models"
	_ "modernc.org/sqlite"
)

// SQLiteStore handles SQLite database operations
type SQLiteStore struct {
	DB *sql.DB
}

// NewSQLiteStore creates a new SQLite store and initializes schema
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Initialize schema
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS raw (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		source TEXT NOT NULL,
		payload BLOB NOT NULL
	);

	CREATE TABLE IF NOT EXISTS snapshot (
		ts TEXT PRIMARY KEY,
		location TEXT NOT NULL,

		-- Weather (OpenMeteo)
		temp_c REAL,
		humidity REAL,
		wind REAL,
		precip REAL,
		cloud_cover REAL,
		visibility_km REAL,

		-- Environment / Air Quality (OpenAQ)
		pm25 REAL,
		pm10 REAL,
		ozone REAL,
		no2 REAL,
		so2 REAL,
		co REAL,

		-- Mobility (HERE, OpenSky, Movebank)
		traffic_speed_kmh REAL,
		traffic_jam_factor REAL,
		flight_count INTEGER,
		avg_altitude_m REAL,
		active_species INTEGER,
		animals_tracked INTEGER,
		avg_migration_pace_km_day REAL,

		-- Finance (AlphaVantage, NASDAQ)
		stock_price REAL,
		stock_symbol TEXT,
		commodity_price REAL,
		commodity_symbol TEXT,
		market_cap REAL,
		volume INTEGER,
		nasdaq_index REAL,
		volume_traded BIGINT,

		-- Energy (Grid, US Energy Info, Ember)
		electricity_price_usd REAL,
		generation_mwh REAL,
		renewable_percent REAL,
		grid_load REAL,
		carbon_intensity_gco2_kwh REAL,
		grid_utilization_percent REAL,
		natural_gas_price_mmbtu REAL,
		coal_percent REAL,
		gas_percent REAL,
		nuclear_percent REAL,

		-- Health (CDC FluView)
		flu_cases INTEGER,
		ili_percent REAL,
		hospital_admissions INTEGER,

		-- Agriculture (USDA NASS)
		crop_yield REAL,
		crop_type TEXT,
		soil_moisture_percent REAL,
		precip_forecast_mm REAL,
		production_bushels REAL,
		price_per_bushel REAL,
		harvested_acres REAL,

		-- Disasters (FEMA)
		active_disasters INTEGER,
		disaster_type TEXT,
		severity INTEGER,
		affected_counties INTEGER
	);

	CREATE TABLE IF NOT EXISTS semantic_record (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		location TEXT NOT NULL,
		ts TEXT NOT NULL,
		category TEXT NOT NULL,
		summary TEXT NOT NULL,
		snapshot_ts TEXT,
		FOREIGN KEY (snapshot_ts) REFERENCES snapshot(ts)
	);

	CREATE INDEX IF NOT EXISTS idx_semantic_location_ts ON semantic_record(location, ts);
	CREATE INDEX IF NOT EXISTS idx_semantic_category ON semantic_record(category);

	CREATE TABLE IF NOT EXISTS agg_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		location TEXT NOT NULL,
		metric TEXT NOT NULL,
		timeframe TEXT NOT NULL,
		value REAL NOT NULL,
		computed_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_agg_location_metric ON agg_metrics(location, metric, timeframe);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		location TEXT NOT NULL,
		ts TEXT NOT NULL,
		event_type TEXT NOT NULL,
		severity REAL NOT NULL,
		description TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_events_location_ts ON events(location, ts);

	-- Embeddings: store vector as JSON text for portability
	CREATE TABLE IF NOT EXISTS snapshot_embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_ts TEXT NOT NULL,
		location TEXT NOT NULL,
		summary TEXT NOT NULL,
		embedding TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (snapshot_ts) REFERENCES snapshot(ts)
	);
	CREATE INDEX IF NOT EXISTS idx_embeddings_location_ts ON snapshot_embeddings(location, snapshot_ts);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &SQLiteStore{DB: db}, nil
}

// InsertSnapshot persists a unified snapshot to the database
func (s *SQLiteStore) InsertSnapshot(snap models.Snapshot) error {
	placeholder := strings.Repeat("?,", 52) + "?" // 53 placeholders for 53 columns

	sql := fmt.Sprintf(`INSERT INTO snapshot
		(ts, location,
		 temp_c, humidity, wind, precip, cloud_cover, visibility_km,
		 pm25, pm10, ozone, no2, so2, co,
		 traffic_speed_kmh, traffic_jam_factor, flight_count, avg_altitude_m, active_species, animals_tracked, avg_migration_pace_km_day,
		 stock_price, stock_symbol, commodity_price, commodity_symbol, market_cap, volume, nasdaq_index, volume_traded,
		 electricity_price_usd, generation_mwh, renewable_percent, grid_load, carbon_intensity_gco2_kwh, grid_utilization_percent, natural_gas_price_mmbtu, coal_percent, gas_percent, nuclear_percent,
		 flu_cases, ili_percent, hospital_admissions,
		 crop_yield, crop_type, soil_moisture_percent, precip_forecast_mm, production_bushels, price_per_bushel, harvested_acres,
		 active_disasters, disaster_type, severity, affected_counties)
		VALUES (%s)`, placeholder)

	_, err := s.DB.Exec(
		sql,
		snap.Timestamp.Format(time.RFC3339),
		snap.Location,

		snap.Weather.TemperatureC,
		snap.Weather.Humidity,
		snap.Weather.WindSpeedMS,
		snap.Weather.PrecipMM,
		snap.Weather.CloudCover,
		snap.Weather.Visibility,

		snap.Environment.PM25,
		snap.Environment.PM10,
		snap.Environment.Ozone,
		snap.Environment.NO2,
		snap.Environment.SO2,
		snap.Environment.CO,

		snap.Mobility.TrafficSpeedKmH,
		snap.Mobility.TrafficJamFactor,
		snap.Mobility.FlightCount,
		snap.Mobility.AvgAltitudeM,
		snap.Mobility.ActiveSpecies,
		snap.Mobility.AnimalsTracked,
		snap.Mobility.AvgMigrationPaceKMDay,

		snap.Finance.StockPrice,
		snap.Finance.StockSymbol,
		snap.Finance.CommodityPrice,
		snap.Finance.CommoditySymbol,
		snap.Finance.MarketCap,
		snap.Finance.Volume,
		snap.Finance.NASDAQIndex,
		snap.Finance.VolumeTraded,

		snap.Energy.ElectricityPriceUSD,
		snap.Energy.GenerationMWh,
		snap.Energy.RenewablePercent,
		snap.Energy.GridLoad,
		snap.Energy.CarbonIntensity,
		snap.Energy.GridUtilizationPercent,
		snap.Energy.NaturalGasPriceMmbtu,
		snap.Energy.CoalPercent,
		snap.Energy.GasPercent,
		snap.Energy.NuclearPercent,

		snap.Health.FluCases,
		snap.Health.ILIPercent,
		snap.Health.HospitalAdmissions,

		snap.Agriculture.CropYield,
		snap.Agriculture.CropType,
		snap.Agriculture.SoilMoisture,
		snap.Agriculture.PrecipForecast,
		snap.Agriculture.ProductionBushels,
		snap.Agriculture.PricePerBushel,
		snap.Agriculture.HarvestedAcres,

		snap.Disasters.ActiveDisasters,
		snap.Disasters.DisasterType,
		snap.Disasters.Severity,
		snap.Disasters.AffectedCounties,
	)

	return err
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}
