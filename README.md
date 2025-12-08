# EdgeSight MVP - Phase 1 Complete

## Architecture Overview

```
EdgeSight/
├── go-ingest/          # Data ingestion & REST API
│   ├── cmd/
│   │   ├── ingest/     # Data collection service
│   │   └── api/        # REST API server
│   ├── internal/
│   │   ├── clients/    # API clients (OpenAQ, NASDAQ, EIA, etc.)
│   │   ├── models/     # Data models
│   │   ├── store/      # SQLite persistence
│   │   ├── canonicalizer/  # Data unification
│   │   └── semantic/   # LLM-ready summaries (for Phase 2)
│   └── edgesight.db    # SQLite database
│
└── edgesight-ui/       # Frontend dashboard
    ├── index.html
    ├── app.js
    └── styles.css
```

## Quick Start

### 1. Collect Data (Run Ingestion Service)

```bash
cd go-ingest

# Set API keys (optional - some services work without keys)
$env:OPENAQ_API_KEY="your-key-here"
$env:ALPHAVANTAGE_API_KEY="your-key-here"
$env:NASDAQ_API_KEY="your-key-here"
$env:EIA_API_KEY="your-key-here"
$env:NASS_API_KEY="your-key-here"

# Run data ingestion (this populates edgesight.db)
.\bin\ingest.exe
```

### 2. Start REST API Server

```bash
cd go-ingest

# Start API on port 8080 (default)
.\bin\api.exe

# Or specify custom port
$env:API_PORT="3000"
.\bin\api.exe
```

### 3. Launch Frontend Dashboard

```bash
cd edgesight-ui

# Option 1: Python simple server
python -m http.server 8000

# Option 2: PHP server
php -S localhost:8000

# Option 3: Node.js http-server (install: npm i -g http-server)
http-server -p 8000

# Then open: http://localhost:8000
```

## API Endpoints

### Health Check
```
GET /health
```

### Get Latest Snapshot
```
GET /api/v1/snapshots/latest?location=Los%20Angeles
```

### Get Snapshots by Time Range
```
GET /api/v1/snapshots/range?location=Los%20Angeles&start=2025-12-07T00:00:00Z&end=2025-12-08T23:59:59Z
```

### Get Recent Snapshots (with pagination)
```
GET /api/v1/snapshots?location=Los%20Angeles&hours=24
```

### Get Metric Time Series
```
GET /api/v1/metrics/series?metric=temp_c&location=Los%20Angeles&start=2025-12-01T00:00:00Z&end=2025-12-08T23:59:59Z
```

Available metrics:
- Weather: `temp_c`, `humidity`, `wind`, `cloud_cover`
- Environment: `pm25`, `pm10`, `ozone`, `no2`, `so2`, `co`
- Energy: `grid_load`, `renewable_percent`, `carbon_intensity_gco2_kwh`
- Finance: `nasdaq_index`, `stock_price`
- Health: `flu_cases`, `ili_percent`, `hospital_admissions`
- Agriculture: `crop_yield`, `price_per_bushel`, `production_bushels`

## Data Sources

### Currently Integrated
1. **OpenMeteo** - Weather data
2. **OpenAQ** - Air quality sensors
3. **AlphaVantage** - Stock prices
4. **NASDAQ Data Link** - Market index
5. **Ember Climate** - Carbon intensity & generation mix
6. **Grid Monitoring** - Power grid status (mock data)
7. **EIA** - US Energy Information Administration
8. **USDA NASS** - Agricultural statistics
9. **FEMA** - Disaster declarations (static JSON)
10. **CDC FluView** - Influenza surveillance
11. **Movebank** - Animal migration tracking

### Removed
- ~~CityBikes~~ (replaced with more relevant energy/ag data)

## Frontend Features

- ✅ Real-time dashboard with auto-refresh (60s)
- ✅ 8 data categories displayed
- ✅ Responsive design (mobile-friendly)
- ✅ Dark theme UI
- ✅ Location selection
- ✅ Error handling & loading states
- ✅ CORS-enabled for local development

## Next Steps (Phase 2 - LLM Integration)

1. **Vector Database**: Add SQLite vector extension (`sqlite-vec`)
2. **Embeddings**: Generate embeddings from semantic summaries
3. **LLM Service**: Integrate Ollama with Mistral 7B or Llama 3.2
4. **Natural Language Queries**: 
   - "What's the air quality trend in LA?"
   - "How does renewable energy correlate with carbon intensity?"
   - "Show me disaster patterns over the last month"

## Development Notes

### Building
```bash
# Build ingestion service
go build -o bin/ingest.exe cmd/ingest/main.go

# Build API server
go build -o bin/api.exe cmd/api/main.go
```

### Database Schema
SQLite database with single `snapshot` table containing all metrics:
- Timestamp-indexed for fast queries
- 50+ columns covering all data domains
- Supports time-series analysis
- Ready for vector extension (Phase 2)

### API Design Philosophy
- REST with JSON responses
- CORS-enabled for web frontends
- Query parameters for filtering
- RFC3339 timestamps
- Error responses with descriptive messages

## MVP Checklist

- ✅ Data ingestion from 11 sources
- ✅ SQLite persistence with time-series support
- ✅ REST API with 5 endpoints
- ✅ Web dashboard UI
- ✅ Real-time data visualization
- ✅ Responsive design
- ⏳ LLM integration (Phase 2)
- ⏳ Vector search (Phase 2)
- ⏳ Advanced analytics (Phase 2)

## Testing

1. **Ingest data**: Run `ingest.exe` to populate database
2. **Verify database**: Check `edgesight.db` was created
3. **Start API**: Run `api.exe`, should see "EdgeSight API Server starting on port 8080"
4. **Test endpoint**: `curl http://localhost:8080/health`
5. **Launch UI**: Start web server in `edgesight-ui/`
6. **View dashboard**: Open browser to `http://localhost:8000`

## Troubleshooting

**API not starting?**
- Check port 8080 isn't in use
- Set `$env:API_PORT="3001"` to use different port

**No data showing?**
- Run `ingest.exe` first to collect data
- Check `edgesight.db` exists
- Verify API is running (`curl http://localhost:8080/health`)

**CORS errors?**
- API has CORS enabled by default
- Make sure you're accessing frontend via http server (not `file://`)

**Frontend not updating?**
- Check browser console for errors
- Verify API URL in `app.js` matches your API server
- Check network tab for failed requests
