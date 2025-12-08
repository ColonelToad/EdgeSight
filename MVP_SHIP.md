# EdgeSight MVP Ship Checklist

## ✅ Completed Components

### Backend (Go API)
- [x] Data ingestion pipeline (OpenMeteo, OpenAQ, AlphaVantage, FRED, EIA, FEMA, CDC, USDA, Ember, Grid.dev, Movebank)
- [x] MQTT subscriber & simulator
- [x] SQLite persistence with schema
- [x] Snapshot embedding & storage
- [x] Semantic search (cosine similarity)
- [x] `/api/v1/snapshots/*` endpoints (latest, range, all)
- [x] `/api/v1/metrics/series` endpoint
- [x] `/api/v1/search` endpoint (semantic search)
- [x] `/api/v1/query` endpoint (search + LLM answer)
- [x] CORS enabled

### Embeddings Sidecar (Python FastAPI)
- [x] `/embed` endpoint (sentence-transformers all-MiniLM-L6-v2)
- [x] `/query` endpoint (Qwen 1.5B language model)
- [x] Lazy-loaded models (avoids startup time bloat)
- [x] Health check endpoint

### Frontend (.NET 10 + SPA)
- [x] C# ASP.NET Core server on port 5174
- [x] `/api/query` proxy to Go API
- [x] Query UI (location, question inputs)
- [x] Answer display with sources
- [x] Source cards with scores & metadata
- [x] Static file serving (CSS, JS, HTML)
- [x] Responsive design

### Data Sources
- [x] **Weather**: OpenMeteo (temp, humidity, wind, clouds, precipitation)
- [x] **Air Quality**: OpenAQ (PM2.5, PM10, O3, NO2, SO2, CO)
- [x] **Finance**: AlphaVantage (stock prices), FRED (NASDAQ), Stooq (fallback)
- [x] **Energy**: EIA, Ember, Grid.dev (prices, generation, carbon intensity, grid load)
- [x] **Health**: CDC NREVSS (RSV/flu data)
- [x] **Agriculture**: USDA NASS (crop data—via file fallback)
- [x] **Disasters**: FEMA (active disasters, severity)
- [x] **Mobility**: OpenSky (flights), Movebank (animal migration)
- [x] **Real-time**: MQTT (temperature, humidity, power)

## 🚀 Running the MVP

### 1. Start Embedding + LLM Sidecar
```powershell
cd c:\Users\legot\EdgeSight
python embedding_sidecar.py
```
Waits on port 9000.

### 2. Start Go API
```powershell
cd c:\Users\legot\EdgeSight\go-ingest
$env:API_PORT="8090"
$env:EMBEDDING_ENDPOINT="http://localhost:9000"
go run ./cmd/api
```
Serves on port 8090.

### 3. Start .NET Frontend
```powershell
cd c:\Users\legot\EdgeSight\edgesight-ui\EdgeSight.Frontend
$env:EDGE_API_BASE="http://localhost:8090/api/v1"
dotnet run --no-launch-profile --urls http://localhost:5174
```
Opens at http://localhost:5174.

### 4. (Optional) Run Ingestion
```powershell
cd c:\Users\legot\EdgeSight\go-ingest
go run ./cmd/ingest
```

### 5. (Optional) Run MQTT Simulator
```powershell
cd c:\Users\legot\EdgeSight\go-ingest
go run ./cmd/mqtt-sim
```

## 📋 Files Ready for Ship

- [x] `.gitignore` (root + go-ingest) updated
- [x] README.md documented
- [x] QUICKSTART.md provided
- [x] ARCHITECTURE.md describes system
- [x] All source code committed
- [x] No API keys hardcoded (use .env)

## ⚠️ Known Limitations (For Next Phase)

- LLM model (Qwen 1.5B) is small—good for demo, not production
- No scheduler—ingest runs on-demand
- Traffic/flight data: not yet ingested (planned for Phase 3)
- LanceDB: not integrated yet (for chunked historical data)
- Frontend: basic Q&A interface, no time-series graphs
- Embeddings: stored in SQLite, not vector DB (fine for MVP scale)

## 🎯 Next Steps (Phase 3)

1. Rewrite API in Rust (consolidate overhead)
2. Add LanceDB for historical data chunking
3. Implement scheduler for periodic ingestion
4. Frontend: add time-series visualization (Chart.js)
5. Expand MQTT integration (sensor dashboard)
6. Traffic/flight data ingestion & chunking strategy

---

**Status**: Ready to ship ✅
**Date**: 2025-12-08
