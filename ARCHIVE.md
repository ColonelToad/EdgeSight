# EdgeSight - Project Archive

**Archive Date:** January 6, 2026  
**Status:** Learning project completed, archived for future reference

## Why This Project Was Built

EdgeSight was created as a hands-on learning project to:
1. **Learn Go** - First substantial Go project, exploring HTTP APIs, concurrency, and database patterns
2. **Explore semantic search** - Experiment with embeddings, vector similarity, and LLM-based Q&A over environmental data
3. **Prototype polyglot architecture** - Go for orchestration, Python for ML, .NET for frontend

**Original Thesis:** Build a system to answer questions like "How is air quality affecting my location?" by combining environmental, economic, and real-time sensor data.

## Why It's Being Archived

**Mission Accomplished:**
- ✅ Learned Go fundamentals (HTTP routing, SQLite ORM, error handling, package structure)
- ✅ Built working embedding sidecar with LLM integration
- ✅ Implemented semantic search over canonical data snapshots
- ✅ Created end-to-end pipeline: ingest → embed → search → LLM → frontend

**Shift in Focus:**
The original research question has been refocused into a dedicated quantitative/HFT project. Maintaining two similar projects with overlapping themes would dilute effort. EdgeSight served its purpose as an exploration vehicle—time to commit to the primary thesis.

**Key Insight:**
Learning projects don't need to become long-term commitments. EdgeSight was a successful experiment that taught valuable patterns now being applied elsewhere.

## What Was Completed

### Phase 1: Data Ingestion & API ✅
- **Data Sources:** OpenMeteo, OpenAQ, AlphaVantage, FRED, EIA, FEMA, CDC, USDA, Ember, Grid.dev, Movebank
- **Canonicalization:** All sources mapped to unified `DataSnapshot` schema
- **Storage:** SQLite with `snapshot` and `snapshot_embeddings` tables
- **REST API:** Go HTTP server on port 8090 with `/api/v1/query`, `/api/v1/snapshots`, `/api/v1/search`

### Phase 2: Semantic Search & LLM ✅
- **Embedding Service:** Python FastAPI sidecar (port 9000) using sentence-transformers (all-MiniLM-L6-v2)
- **LLM Integration:** Qwen 1.5B via HuggingFace transformers for inference
- **Vector Search:** Cosine similarity in SQLite (MVP-scale; LanceDB planned for Phase 3)
- **Query Pipeline:** User question → embed → search top 5 snapshots → LLM answer with sources

### Phase 2.5: MQTT Real-Time Integration ✅
- **MQTT Broker:** Eclipse Mosquitto (port 1883)
- **Publishers:** Go MQTT client + mqtt-sim for virtual sensors
- **Ingestion:** MQTT messages saved to `snapshot` table with auto-embedding on insert

### Frontend: .NET 10 Web App ✅
- **Stack:** ASP.NET Core minimal API + SPA
- **Features:** Query input, LLM answer display, source cards with similarity scores
- **Ports:** Frontend on 5174, proxies to Go API on 8090

### EdgeSight-IoT Spinoff (Separate Repo) ✅
Standalone MQTT simulator + real-time dashboard:
- Virtual sensors (weather, air quality, power, water) using public APIs
- Flask WebSocket dashboard with live visualization
- Device integration-ready (drop in real sensor publishers)
- **Status:** Scaffolded, not actively maintained

## What Was NOT Completed (Phase 3 Roadmap)

These were planned but deprioritized:

1. **Rust Rewrite** - Consolidate Go API into Rust for performance
2. **LanceDB Integration** - Replace SQLite embeddings with proper vector store
3. **Historical Data Chunking** - Flight/traffic data with route-based partitioning
4. **Ingestion Scheduler** - APScheduler for automated data refresh
5. **Advanced Frontend** - Time-series graphs, MQTT live stream visualization
6. **Production Deployment** - Docker Compose orchestration, monitoring

## Key Learnings & Patterns

### 1. Embedding Sidecar Pattern
**What:** Separate Python service for ML workloads, called via HTTP from Go  
**Why:** Language-appropriate separation (Go for orchestration, Python for transformers)  
**Trade-off:** Network latency (~50ms) vs. cleaner architecture  
**Reusable:** Apply this to financial news embeddings in quant projects

### 2. Canonical Data Schema
**What:** All external APIs mapped to unified `DataSnapshot` model  
**Why:** Simplifies downstream processing (embedding, search, LLM context)  
**Challenge:** Metadata loss (stock price != weather != emissions)  
**Lesson:** Canonical schema works for MVP; specialized schemas needed at scale

### 3. SQLite for Embeddings (MVP-Scale)
**What:** Store 384-dim vectors as JSON, cosine search in Go  
**Performance:** <100ms for 10k snapshots; breaks around 100k+  
**When to Use:** Prototypes, local-first apps, embedded systems  
**When to Upgrade:** Production scale → LanceDB, Pinecone, or Qdrant

### 4. LLM Context Window Management
**What:** Top-5 semantic search results used as RAG context  
**Challenge:** Qwen 1.5B (2k token context) limits how much data you can pass  
**Solution:** Summarize snapshots in canonicalizer before embedding  
**Reusable:** Financial event summarization for trading signal generation

### 5. Go HTTP Patterns
**Learned:**
- Middleware chaining for logging, CORS, auth
- Context propagation for cancellation
- Client timeouts (default HTTP client has no timeout!)
- Error wrapping with `fmt.Errorf("%w", err)`

**Anti-patterns to avoid:**
- Global variables for state (use dependency injection)
- Unchecked JSON marshal errors
- Goroutines without proper cleanup

## Technical Debt & Known Issues

**Not Fixed (Low Priority for Archive):**
1. **Embedding backfill:** Only new snapshots get embedded; need batch job for historical data
2. **MQTT reconnect logic:** Doesn't auto-reconnect on broker restart
3. **Frontend error handling:** No user feedback when API is down
4. **Python sidecar startup:** 10-30s cold start for LLM model load
5. **API key management:** .env files not encrypted; use Vault in production

**Security Notes:**
- API keys in `go-ingest/.env` are personal test accounts with low/no rate limits
- Move to environment variables or secret manager for any deployment
- MQTT broker has no authentication (acceptable for local testing only)

## File Inventory

**Core Components:**
- [go-ingest/](./go-ingest/) - Go API and ingestion service
- [embedding_sidecar.py](./embedding_sidecar.py) - Python ML service
- [edgesight-ui/](./edgesight-ui/) - .NET 10 frontend
- [MVP_SHIP.md](./MVP_SHIP.md) - Complete run instructions and checklist

**Documentation:**
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System design
- [QUICKSTART.md](./QUICKSTART.md) - 5-minute setup guide
- [PHASE1_COMPLETION.md](./PHASE1_COMPLETION.md) - Ingestion milestone
- [PHASE2_PLANNING.md](./PHASE2_PLANNING.md) - Embeddings roadmap

**Data:**
- `edgesight.db` - SQLite database (excluded from git)
- `Respiratory_Syncytial_Virus_Laboratory_Data_(NREVSS)_20251208.csv` - CDC sample data

**Related Projects:**
- [EdgeSight-IoT](https://github.com/ColonelToad/EdgeSight-IoT) - MQTT simulator (separate repo)

## How to Run (For Future Reference)

### Prerequisites
```powershell
# Go 1.21+
go version

# Python 3.10+ with pip
python --version

# Docker (for MQTT broker)
docker --version

# .NET 10 SDK
dotnet --version
```

### 3-Terminal Startup

**Terminal 1: Python Embedding Sidecar**
```powershell
cd C:\Users\legot\EdgeSight
python embedding_sidecar.py
# Wait for "Uvicorn running on http://0.0.0.0:9000"
```

**Terminal 2: Go API**
```powershell
cd C:\Users\legot\EdgeSight\go-ingest
$env:API_PORT="8090"
go run ./cmd/api
```

**Terminal 3: .NET Frontend**
```powershell
cd C:\Users\legot\EdgeSight\edgesight-ui
dotnet run --no-launch-profile --urls http://localhost:5174
```

**Access:** http://localhost:5174

### Optional: MQTT Real-Time Data
```powershell
# Terminal 4: MQTT Broker
docker run -d --name mosquitto -p 1883:1883 eclipse-mosquitto

# Terminal 5: MQTT Simulator
cd C:\Users\legot\EdgeSight\go-ingest
go run .\cmd\mqtt\main.go
```

## Salvageable Components for Future Projects

**High-Value Patterns:**
1. **Embedding sidecar architecture** → Apply to financial news analysis
2. **Semantic search implementation** → Use for backtesting event correlation
3. **Go HTTP API scaffolding** → Template for future Go services
4. **Canonical data pipeline** → Adapt for multi-source market data normalization

**Code to Reuse:**
- `internal/embeddings/client.go` - HTTP client for embedding service
- `internal/store/sqlite.go` - SQLite ORM patterns
- `embedding_sidecar.py` - FastAPI + transformers boilerplate
- `cmd/api/main.go` - Go HTTP server structure

## Acknowledgments

This project was a sandbox for learning. The patterns explored here will inform future work, particularly in quantitative finance where semantic search over market events could provide edge in signal generation.

**Key Resources That Helped:**
- [Sentence Transformers Documentation](https://www.sbert.net/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)

---

**Final Commit:** `469c05c` - "Archive EdgeSight: Learning project complete"  
**Repository:** https://github.com/ColonelToad/EdgeSight  
**Archived By:** ColonelToad  
**Date:** January 6, 2026

*Learning projects are meant to be exploratory, not eternal. Mission accomplished. ✅*
