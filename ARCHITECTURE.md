# EdgeSight Architecture & Deployment

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          FRONTEND TIER                          │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │           Web Dashboard (HTML/CSS/JS)                  │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │  - 8 Data Category Sections                      │  │    │
│  │  │  - 40+ Real-time Metrics                         │  │    │
│  │  │  - Location-based Queries                        │  │    │
│  │  │  - Auto-refresh Every 60s                        │  │    │
│  │  │  - Dark Theme + Responsive                       │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  │              http://localhost:8000                     │    │
│  └────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              │ REST API Calls                   │
│                              ▼                                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               │ HTTP/JSON
                               │
┌──────────────────────────────┴──────────────────────────────────┐
│                          API TIER                               │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │           REST API Server (Go - api.exe)               │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │  GET /health                                     │  │    │
│  │  │  GET /api/v1/snapshots/latest                    │  │    │
│  │  │  GET /api/v1/snapshots/range                     │  │    │
│  │  │  GET /api/v1/snapshots                           │  │    │
│  │  │  GET /api/v1/metrics/series                      │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  │              http://localhost:8080                     │    │
│  └────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              │ SQL Queries                      │
│                              ▼                                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────┴──────────────────────────────────┐
│                      PERSISTENCE TIER                           │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │        SQLite Database (edgesight.db)                  │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │  Table: snapshot                                 │  │    │
│  │  │  - 1 row per timestamp per location             │  │    │
│  │  │  - 50+ columns (all data domains)                │  │    │
│  │  │  - Time-indexed for fast range queries           │  │    │
│  │  │  - Ready for vector extension (Phase 2)          │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  └────────────────────────────────────────────────────────┘    │
│                              ▲                                  │
│                              │ Insert/Update                    │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┴──────────────────────────────────┐
│                    INGESTION TIER                               │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │      Data Ingestion Service (Go - ingest.exe)          │    │
│  │  ┌──────────────────────────────────────────────────┐  │    │
│  │  │ 1. Fetch from API Sources (parallel)             │  │    │
│  │  │ 2. Parse Responses                               │  │    │
│  │  │ 3. Normalize Data Format                          │  │    │
│  │  │ 4. Unify via Canonicalizer                        │  │    │
│  │  │ 5. Persist to Database                            │  │    │
│  │  └──────────────────────────────────────────────────┘  │    │
│  └────────────────────────────────────────────────────────┘    │
│                              ▲                                  │
│                 REST API Calls (11 sources)                    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
        ┌──────────────────────┴──────────────────────┐
        │                                             │
   ┌────▼────┐  ┌──────────┐  ┌──────────┐  ┌──────▼─┐
   │OpenMeteo │  │ OpenAQ   │  │Ember     │  │AlphaV  │
   │(Weather) │  │(AirQual) │  │(Carbon)  │  │(Stock) │
   └────┬────┘  └──────────┘  └──────────┘  └──────┬─┘
        │
   ┌────▼────┐  ┌──────────┐  ┌──────────┐  ┌──────▼─┐
   │NASDAQ    │  │Grid      │  │EIA       │  │NASS    │
   │(Index)   │  │(Load)    │  │(Energy)  │  │(Crops) │
   └────┬────┘  └──────────┘  └──────────┘  └──────┬─┘
        │
   ┌────▼────┐  ┌──────────┐  ┌──────────┐
   │FEMA      │  │CDC       │  │Movebank  │
   │(Disaster)│  │(Health)  │  │(Wildlife)│
   └─────────┘  └──────────┘  └──────────┘
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DATA COLLECTION                             │
│                                                                     │
│  11 API Sources                                                    │
│  (Weather, Energy, Finance, Health, Disaster, Agriculture, etc)   │
│                           │                                        │
│                           ▼                                        │
│                   ┌───────────────┐                               │
│                   │  API Clients  │ (internal/clients/)           │
│                   │  (11 structs) │                               │
│                   │               │                               │
│                   │ • OpenMeteo   │                               │
│                   │ • OpenAQ      │                               │
│                   │ • Ember       │                               │
│                   │ • Grid        │                               │
│                   │ • EIA         │                               │
│                   │ • NASS        │                               │
│                   │ • NASDAQ      │                               │
│                   │ • AlphaVantage│                               │
│                   │ • FEMA        │                               │
│                   │ • CDC         │                               │
│                   │ • Movebank    │                               │
│                   └───────────────┘                               │
│                           │                                        │
│                           │ Different schemas                      │
│                           │ Different formats                      │
│                           │ Different types                        │
│                           ▼                                        │
│                  ┌──────────────────┐                            │
│                  │  Canonicalizer   │ (internal/canonicalizer/)  │
│                  │                  │                             │
│                  │ Unified Model:   │                             │
│                  │ ┌──────────────┐ │                             │
│                  │ │ Snapshot     │ │                             │
│                  │ │ ├─ Weather   │ │                             │
│                  │ │ ├─ Environment
│                  │ │ ├─ Mobility  │ │                             │
│                  │ │ ├─ Finance   │ │                             │
│                  │ │ ├─ Energy    │ │                             │
│                  │ │ ├─ Health    │ │                             │
│                  │ │ ├─ Agriculture
│                  │ │ └─ Disasters │ │                             │
│                  │ └──────────────┘ │                             │
│                  └──────────────────┘                            │
│                           │                                        │
│                           │ Unified structure                      │
│                           ▼                                        │
│                    ┌─────────────┐                               │
│                    │   SQLite    │                               │
│                    │  Database   │ (edgesight.db)               │
│                    │             │                               │
│                    │ snapshot    │                               │
│                    │ table with  │                               │
│                    │ 50+ cols    │                               │
│                    │ indexed on  │                               │
│                    │ timestamp   │                               │
│                    └─────────────┘                               │
│                           │                                        │
└───────────────────────────┼────────────────────────────────────────┘
                            │
┌───────────────────────────┼────────────────────────────────────────┐
│                         DATA SERVING                               │
│                                                                    │
│                           ▼                                        │
│                    ┌─────────────┐                               │
│                    │  REST API   │ (cmd/api/)                   │
│                    │  Server     │                               │
│                    │             │                               │
│                    │ Endpoints:  │                               │
│                    │ • /health   │                               │
│                    │ • /latest   │                               │
│                    │ • /range    │                               │
│                    │ • /series   │                               │
│                    └─────────────┘                               │
│                           │                                        │
│                           │ JSON over HTTP                         │
│                           ▼                                        │
│                   ┌──────────────┐                               │
│                   │   Browser    │                               │
│                   │  Dashboard   │ (edgesight-ui/)              │
│                   │              │                               │
│                   │ Displays:    │                               │
│                   │ • 8 sections │                               │
│                   │ • 40+ cards  │                               │
│                   │ • Live data  │                               │
│                   └──────────────┘                               │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

## Deployment Topology

```
┌──────────────────────────────────────────────────────────┐
│                   DEVELOPMENT MACHINE                    │
│                   (Windows/Mac/Linux)                    │
│                                                          │
│  ┌────────────────────┐  ┌────────────────────┐         │
│  │  Browser           │  │  Terminal Windows  │         │
│  │                    │  │                    │         │
│  │ localhost:8000     │  │ Terminal 1: ingest │         │
│  │ ├─ Dashboard       │  │ Terminal 2: api    │         │
│  │ └─ Real-time data  │  │ Terminal 3: serve  │         │
│  │                    │  │                    │         │
│  └────────┬───────────┘  └────────┬───────────┘         │
│           │                       │                      │
│           └───────────┬───────────┘                      │
│                       │ HTTP                             │
│          ┌────────────┴──────────┐                       │
│          │                       │                       │
│      ┌───▼────┐            ┌─────▼────┐                 │
│      │ api    │            │  http    │                 │
│      │:8080   │            │ :8000    │                 │
│      └───┬────┘            └──────────┘                 │
│          │                                               │
│      ┌───▼────────┐                                      │
│      │edgesight.db│                                      │
│      │  (SQLite)  │                                      │
│      └────────────┘                                      │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

## Service Dependencies

```
edgesight-ui (Frontend)
└─→ http://localhost:8080
    │
    api.exe (REST Server)
    └─→ edgesight.db (SQLite)
        └─→ Data from ingest.exe

ingest.exe (Data Collection)
└─→ (runs periodically or once)
    ├─→ 11 REST APIs
    └─→ edgesight.db (write)
```

## File Structure with Responsibilities

```
EdgeSight/
│
├── go-ingest/                    # Backend service
│   │
│   ├── cmd/
│   │   ├── ingest/
│   │   │   └── main.go           # Orchestrates data collection
│   │   │                          # Calls all API clients
│   │   │                          # Runs canonicalizer
│   │   │                          # Persists to database
│   │   │
│   │   └── api/
│   │       └── main.go           # REST API server
│   │                              # Routes requests
│   │                              # Serves data from DB
│   │
│   ├── internal/clients/          # Data source integrations
│   │   ├── openmeteo.go          # Weather data
│   │   ├── openaq.go             # Air quality
│   │   ├── ember.go              # Carbon intensity
│   │   ├── grid.go               # Grid load/status
│   │   ├── eia.go                # Energy stats
│   │   ├── nass.go               # Agricultural data
│   │   ├── nasdaq.go             # Market index
│   │   ├── alphavantage.go       # Stock prices
│   │   ├── fema.go               # Disaster data
│   │   ├── cdc_fluview.go        # Health data
│   │   └── movebank.go           # Wildlife migration
│   │
│   ├── internal/models/
│   │   └── canonical.go          # Unified Snapshot struct
│   │                              # 8 domain types
│   │                              # JSON marshaling
│   │
│   ├── internal/store/
│   │   ├── sqlite.go             # Database initialization
│   │   │                          # Schema definition
│   │   └── queries.go            # SQL operations
│   │                              # SELECT queries
│   │                              # INSERT/UPDATE
│   │
│   ├── internal/canonicalizer/
│   │   └── canonicalizer.go      # BuildSnapshot function
│   │                              # Maps clients → model
│   │
│   ├── bin/
│   │   ├── ingest.exe            # Compiled ingestion service
│   │   └── api.exe               # Compiled API server
│   │
│   └── edgesight.db              # SQLite database
│
├── edgesight-ui/                  # Frontend application
│   ├── index.html                 # Dashboard layout
│   │                              # 8 data sections
│   │                              # 40+ metric cards
│   │
│   ├── app.js                     # Client-side logic
│   │                              # API calls (fetch)
│   │                              # DOM updates
│   │                              # Auto-refresh timer
│   │
│   ├── styles.css                 # Visual styling
│   │                              # Dark theme
│   │                              # Responsive design
│   │
│   └── start.bat                  # Quick launcher script
│
└── Documentation
    ├── INDEX.md                   # Project overview
    ├── README.md                  # Full technical docs
    ├── QUICKSTART.md              # 5-minute setup
    └── PHASE1_SUMMARY.md          # Architecture & design
```

## Technology Stack

```
Backend:
  Language:     Go 1.21+
  Database:     SQLite (modernc.org/sqlite)
  API:          net/http (Go stdlib)
  JSON:         encoding/json (Go stdlib)
  Build:        go build (no external tools)

Frontend:
  Language:     HTML5 / CSS3 / JavaScript (ES6)
  Framework:    None (vanilla)
  Build Tool:   None (serve with http.server)
  Dependencies: Zero external JS libs

Total Binaries: 2 (~15 MB each)
Total Dependencies: Zero external packages
```

## Scaling Considerations

```
Current MVP:
├─ One ingestion run per session
├─ Real-time dashboard (60s refresh)
├─ Single-threaded API
├─ SQLite single-writer
└─ Embedded-device friendly

Future Scaling:
├─ Scheduled ingestion (cron/systemd)
├─ WebSocket real-time updates
├─ Horizontal API servers
├─ PostgreSQL for multi-writer
├─ Redis cache for hot data
└─ Distributed time-series DB (InfluxDB/TimescaleDB)
```

## Security Notes

```
Current MVP:
✓ API endpoints are public (no auth)
✓ All data is read-only
✓ No sensitive information
✓ Local deployment only

For Production:
□ Add API key authentication
□ Rate limiting on endpoints
□ HTTPS/TLS encryption
□ Input validation
□ SQL injection prevention (parametrized queries ✓)
□ CORS restrictions (currently wildcard)
```

---

**For detailed setup instructions, see [QUICKSTART.md](QUICKSTART.md)**

**For architectural deep-dive, see [PHASE1_SUMMARY.md](PHASE1_SUMMARY.md)**
