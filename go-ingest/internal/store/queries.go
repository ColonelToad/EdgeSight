package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ColonelToad/EdgeSight/go-ingest/internal/models"
)

// TimeSeriesPoint represents a single metric value at a point in time
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// SQL column list for SELECT queries
const snapshotColumns = `ts, location,
	temp_c, humidity, wind, precip, cloud_cover, visibility_km,
	pm25, pm10, ozone, no2, so2, co,
	traffic_speed_kmh, traffic_jam_factor, flight_count, avg_altitude_m, active_species, animals_tracked, avg_migration_pace_km_day,
	stock_price, stock_symbol, commodity_price, commodity_symbol, market_cap, volume, nasdaq_index, volume_traded,
	electricity_price_usd, generation_mwh, renewable_percent, grid_load, carbon_intensity_gco2_kwh, grid_utilization_percent, natural_gas_price_mmbtu, coal_percent, gas_percent, nuclear_percent,
	flu_cases, ili_percent, hospital_admissions,
	crop_yield, crop_type, soil_moisture_percent, precip_forecast_mm, production_bushels, price_per_bushel, harvested_acres,
	active_disasters, disaster_type, severity, affected_counties`

// GetLatestSnapshot retrieves the most recent snapshot for a location
func (s *SQLiteStore) GetLatestSnapshot(location string) (*models.Snapshot, error) {
	query := fmt.Sprintf(`SELECT %s FROM snapshot WHERE location = ? ORDER BY ts DESC LIMIT 1`, snapshotColumns)

	row := s.DB.QueryRow(query, location)
	snap, err := scanSnapshot(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no snapshots found for location: %s", location)
	}
	return snap, err
}

// GetSnapshotsByTimeRange retrieves all snapshots for a location within a time range
func (s *SQLiteStore) GetSnapshotsByTimeRange(location string, start, end time.Time) ([]models.Snapshot, error) {
	query := fmt.Sprintf(`SELECT %s FROM snapshot 
	          WHERE location = ? AND ts >= ? AND ts <= ? 
	          ORDER BY ts ASC`, snapshotColumns)

	rows, err := s.DB.Query(query, location, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.Snapshot
	for rows.Next() {
		snap, err := scanSnapshotRow(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *snap)
	}

	return snapshots, rows.Err()
}

// GetMetricSeries retrieves a time series for a specific metric
func (s *SQLiteStore) GetMetricSeries(metric, location string, start, end time.Time) ([]TimeSeriesPoint, error) {
	query := fmt.Sprintf(`SELECT ts, %s FROM snapshot 
	                      WHERE location = ? AND ts >= ? AND ts <= ? AND %s IS NOT NULL
	                      ORDER BY ts ASC`, metric, metric)

	rows, err := s.DB.Query(query, location, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []TimeSeriesPoint
	for rows.Next() {
		var tsStr string
		var value float64
		if err := rows.Scan(&tsStr, &value); err != nil {
			return nil, err
		}

		ts, err := time.Parse(time.RFC3339, tsStr)
		if err != nil {
			return nil, err
		}

		series = append(series, TimeSeriesPoint{
			Timestamp: ts,
			Value:     value,
		})
	}

	return series, rows.Err()
}

// scanSnapshot scans a single row into a Snapshot
func scanSnapshot(row *sql.Row) (*models.Snapshot, error) {
	var snap models.Snapshot
	var tsStr string

	err := row.Scan(
		&tsStr, &snap.Location,
		&snap.Weather.TemperatureC, &snap.Weather.Humidity, &snap.Weather.WindSpeedMS, &snap.Weather.PrecipMM, &snap.Weather.CloudCover, &snap.Weather.Visibility,
		&snap.Environment.PM25, &snap.Environment.PM10, &snap.Environment.Ozone, &snap.Environment.NO2, &snap.Environment.SO2, &snap.Environment.CO,
		&snap.Mobility.TrafficSpeedKmH, &snap.Mobility.TrafficJamFactor, &snap.Mobility.FlightCount, &snap.Mobility.AvgAltitudeM, &snap.Mobility.ActiveSpecies, &snap.Mobility.AnimalsTracked, &snap.Mobility.AvgMigrationPaceKMDay,
		&snap.Finance.StockPrice, &snap.Finance.StockSymbol, &snap.Finance.CommodityPrice, &snap.Finance.CommoditySymbol, &snap.Finance.MarketCap, &snap.Finance.Volume, &snap.Finance.NASDAQIndex, &snap.Finance.VolumeTraded,
		&snap.Energy.ElectricityPriceUSD, &snap.Energy.GenerationMWh, &snap.Energy.RenewablePercent, &snap.Energy.GridLoad, &snap.Energy.CarbonIntensity, &snap.Energy.GridUtilizationPercent, &snap.Energy.NaturalGasPriceMmbtu, &snap.Energy.CoalPercent, &snap.Energy.GasPercent, &snap.Energy.NuclearPercent,
		&snap.Health.FluCases, &snap.Health.ILIPercent, &snap.Health.HospitalAdmissions,
		&snap.Agriculture.CropYield, &snap.Agriculture.CropType, &snap.Agriculture.SoilMoisture, &snap.Agriculture.PrecipForecast, &snap.Agriculture.ProductionBushels, &snap.Agriculture.PricePerBushel, &snap.Agriculture.HarvestedAcres,
		&snap.Disasters.ActiveDisasters, &snap.Disasters.DisasterType, &snap.Disasters.Severity, &snap.Disasters.AffectedCounties,
	)

	if err != nil {
		return nil, err
	}

	snap.Timestamp, err = time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}

// scanSnapshotRow scans a Rows iterator into a Snapshot
func scanSnapshotRow(rows *sql.Rows) (*models.Snapshot, error) {
	var snap models.Snapshot
	var tsStr string

	err := rows.Scan(
		&tsStr, &snap.Location,
		&snap.Weather.TemperatureC, &snap.Weather.Humidity, &snap.Weather.WindSpeedMS, &snap.Weather.PrecipMM, &snap.Weather.CloudCover, &snap.Weather.Visibility,
		&snap.Environment.PM25, &snap.Environment.PM10, &snap.Environment.Ozone, &snap.Environment.NO2, &snap.Environment.SO2, &snap.Environment.CO,
		&snap.Mobility.TrafficSpeedKmH, &snap.Mobility.TrafficJamFactor, &snap.Mobility.FlightCount, &snap.Mobility.AvgAltitudeM, &snap.Mobility.ActiveSpecies, &snap.Mobility.AnimalsTracked, &snap.Mobility.AvgMigrationPaceKMDay,
		&snap.Finance.StockPrice, &snap.Finance.StockSymbol, &snap.Finance.CommodityPrice, &snap.Finance.CommoditySymbol, &snap.Finance.MarketCap, &snap.Finance.Volume, &snap.Finance.NASDAQIndex, &snap.Finance.VolumeTraded,
		&snap.Energy.ElectricityPriceUSD, &snap.Energy.GenerationMWh, &snap.Energy.RenewablePercent, &snap.Energy.GridLoad, &snap.Energy.CarbonIntensity, &snap.Energy.GridUtilizationPercent, &snap.Energy.NaturalGasPriceMmbtu, &snap.Energy.CoalPercent, &snap.Energy.GasPercent, &snap.Energy.NuclearPercent,
		&snap.Health.FluCases, &snap.Health.ILIPercent, &snap.Health.HospitalAdmissions,
		&snap.Agriculture.CropYield, &snap.Agriculture.CropType, &snap.Agriculture.SoilMoisture, &snap.Agriculture.PrecipForecast, &snap.Agriculture.ProductionBushels, &snap.Agriculture.PricePerBushel, &snap.Agriculture.HarvestedAcres,
		&snap.Disasters.ActiveDisasters, &snap.Disasters.DisasterType, &snap.Disasters.Severity, &snap.Disasters.AffectedCounties,
	)

	if err != nil {
		return nil, err
	}

	snap.Timestamp, err = time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}
