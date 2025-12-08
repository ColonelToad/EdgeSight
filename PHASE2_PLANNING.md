# Phase 2 Planning - LLM Integration & MQTT Sensor Simulation

**Date:** December 8, 2025  
**Status:** Planning & Architecture  
**Target:** Hardware simulation with MQTT + local LLM  

---

## 1. CDC FluView Issue - Detailed Analysis

### The Problem
The CDC FluView client is hitting a **405 Method Not Allowed** error when trying to fetch data.

**Root Cause:** The endpoint `https://gis.cdc.gov/grasp/flu2/PostPhase02DataDownload` is designed for **POST requests with form data**, not GET requests.

### Current Implementation (Broken)
```go
// In cdc_fluview.go - This fails
url := fmt.Sprintf("%s/PostPhase02DataDownload?llILIActivityID=-1&llSeasonID=58&llRegionID=12&llGroupID=0", c.baseURL)
resp, err := c.httpCli.Get(url)  // ❌ GET request on POST-only endpoint
```

### Why It's Not Working
The CDC endpoint expects:
1. **POST request** (not GET)
2. **Form-encoded body** with parameters
3. **Specific parameter format** (the query string approach doesn't work)
4. **Session/referer headers** (the endpoint is web-scraping protected)

### Solutions (In Order of Preference)

**Option A: Use CDC's Official JSON Endpoint (BEST)**
The CDC provides a JSON API at `https://gis.cdc.gov/grasp/data.json` for some regions.
```go
// Better approach - Use the public JSON endpoint
url := "https://gis.cdc.gov/grasp/data.json"
resp, err := c.httpCli.Get(url)
// Parse the JSON array for current season data
```

**Option B: Use NREVSS (CDC's Alternative)**
CDC provides NREVSS (National Respiratory and Enteric Virus Surveillance System) data:
```go
// NREVSS endpoint is more stable
url := "https://data.cdc.gov/api/views/a8yy-6fxy/rows.json?accessType=DOWNLOAD"
resp, err := c.httpCli.Get(url)
```

**Option C: Web Scraping with Proper POST (WORKAROUND)**
Implement POST with form data:
```go
data := url.Values{}
data.Set("llILIActivityID", "-1")
data.Set("llSeasonID", "58")
data.Set("llRegionID", "12")

req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// Add referer and user-agent
req.Header.Set("Referer", "https://gis.cdc.gov/grasp/flu/")
req.Header.Set("User-Agent", "EdgeSight/1.0")
```

**Option D: Skip CDC for Now (PRAGMATIC)**
Keep the fallback behavior (graceful failure) and focus on other integrations.
The system already handles errors gracefully with nil checks.

### Recommendation
**Use Option A + Option D:** Try the JSON endpoint, and if it fails, gracefully return nil (current behavior). This requires minimal code change and respects CDC's intended API design.

---

## 2. API Key Management - Updated

### Current Status
You've provided keys for:
- ✅ EIA
- ✅ NASDAQ  
- ✅ NASS
- ✅ OpenAQ
- ✅ AlphaVantage
- ✅ GRID
- ✅ Ember

### To Complete
1. **Movebank** - Add credentials to `.env`:
   ```
   MOVEBANK_USER=your_username
   MOVEBANK_PASS=your_password
   ```
   Update the client to read these and use HTTP Basic Auth.

2. **OpenMeteo** - No key needed (public API)

3. **FEMA** - No key needed (public API)

### `.env` File Structure
```bash
# Finance
ALPHAVANTAGE_API_KEY=Y95YDKCQGOKFE75O
NASDAQ_API_KEY=fwN1zxc6ign_1WUAF7Nn

# Energy
EIA_API_KEY=5UOc0J5WkPqAlOadrcXWQqSepXvJ9UV9Hi7qhMv0
EMBER_API_KEY=9639a452-95bc-bf54-c3d3-6d37284ceab1
GRID_API_KEY=6e79f27cfbfa42e6ada677d1d9a06b65

# Agriculture
NASS_API_KEY=5F265E56-AF2D-3425-9848-0C1A0131DF17

# Environment
OPENAQ_API_KEY=4969b2feccae0d98440d854a10feffcfec6258fb1e3ab6e12c1be211b30e0384

# Wildlife (add these)
MOVEBANK_USER=your_email@example.com
MOVEBANK_PASS=your_password
```

---

## 3. Phase 2 Architecture - LLM + MQTT Integration

### Phase 2 Goals
Transform EdgeSight from **data warehouse** → **intelligent data assistant**

```
Phase 1: Data Collection → Database → Dashboard
                                          ↓
Phase 2: ... → Database → LLM Query Engine → Natural Language Responses
                           ↑
                    Vector Embeddings
```

### 3.1 Local LLM Setup

**Your Choice:** Qwen 2.5 7B (excellent choice!)

#### Why Qwen 2.5 7B is Good for EdgeSight
- ✅ Small (7B = ~14-16GB RAM required, runs on consumer hardware)
- ✅ Fast inference (~100-200ms per query)
- ✅ Good instruction following
- ✅ Multilingual support
- ✅ No API costs
- ✅ Can run offline

#### Integration Without Ollama
Since you already have the GGUF file, load it directly in Go using `go-llama.cpp` bindings:

```go
// New file: internal/llm/engine.go
package llm

import (
	"github.com/go-skynet/go-llama.cpp"
)

type LLMEngine struct {
	model *llama.LLama
}

func NewLLMEngine(modelPath string) (*LLMEngine, error) {
	opts := []llama.PredictOption{
		llama.SetThreads(4),
		llama.SetGrammar(nil),
		llama.SetTopK(50),
		llama.SetTopP(0.95),
		llama.SetTemperature(0.8),
	}
	
	model, err := llama.New(modelPath, opts...)
	if err != nil {
		return nil, err
	}
	
	return &LLMEngine{model: model}, nil
}

func (e *LLMEngine) Query(prompt string) (string, error) {
	result, err := e.model.Predict(prompt, 
		llama.SetTokens(2048),
	)
	return result, err
}
```

**File Structure:**
```
go-ingest/
├── models/
│   └── qwen2.5-7b.gguf          (5-16GB file)
├── internal/
│   └── llm/
│       ├── engine.go            (LLM inference)
│       ├── embedder.go          (Vector generation)
│       └── retriever.go         (Vector search)
```

---

### 3.2 Vector Database & Embeddings

**Data Flow:**
```
Snapshot in DB → Semantic Summary → Embedding Vector → Vector DB
                                          ↓
                          Query → Find Similar Vectors → LLM Context
```

#### Option A: SQLite with Vector Extension (RECOMMENDED)
Use `sqlite-vec` or `sqlite-vss`:

```sql
-- Add vector table to schema
CREATE TABLE snapshot_embeddings (
    id TEXT PRIMARY KEY,
    snapshot_id INTEGER,
    text_summary TEXT,      -- e.g., "High temperature 95F, AQI 156 unhealthy"
    embedding VECTOR(768),  -- or 384 for smaller models
    created_at TIMESTAMP
);

-- Vector search query
SELECT text_summary, distance 
FROM snapshot_embeddings 
WHERE embedding MATCH ?
ORDER BY distance LIMIT 5;
```

**Advantages:**
- ✅ No external service
- ✅ Stays in SQLite (simpler ops)
- ✅ GGUF embedding models (local)
- ✅ Fast (~10ms per search)

#### Option B: LanceDB (if you want to try it)
LanceDB is newer and optimized for vectors, but adds a dependency.

#### Embedding Model Choice
Use a lightweight embedding model:
```
sentence-transformers/all-MiniLM-L6-v2  (384-dim, 22MB)
or
BAAI/bge-small-en-v1.5                  (384-dim, 33MB)
```

Both can run locally without additional services.

---

### 3.3 Query Interface

**New REST Endpoints for Phase 2:**

```go
// GET /api/v1/query?q=what+was+the+highest+temperature+today
// Returns: {"answer": "...", "sources": [...], "confidence": 0.92}

func (s *APIServer) HandleQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	// 1. Generate embedding for query
	queryEmbedding, _ := embedder.Embed(query)
	
	// 2. Search vector DB for context
	relevantSnapshots := store.SearchEmbeddings(queryEmbedding, topK=5)
	
	// 3. Build LLM prompt with context
	context := buildContext(relevantSnapshots)
	prompt := fmt.Sprintf(
		`You are a data analyst answering questions about environmental and market data.
Context: %s
Question: %s
Answer:`,
		context, query,
	)
	
	// 4. Call LLM
	answer, _ := llm.Query(prompt)
	
	// 5. Return response
	json.NewEncoder(w).Encode(QueryResponse{
		Answer:  answer,
		Sources: extractSources(relevantSnapshots),
	})
}
```

**Frontend Enhancement:**
```html
<!-- Add query box to dashboard -->
<div id="query-section">
	<input type="text" id="queryInput" placeholder="Ask about your data...">
	<button onclick="submitQuery()">Ask</button>
	<div id="queryResponse"><!-- LLM response appears here --></div>
</div>
```

---

## 4. MQTT Sensor Simulation - Hardware Constraint Environment

### Your Idea (Excellent!)
```
MQTT Broker ← Simulated Sensors (noisy data)
     ↓
Go Backend (ingests via MQTT)
     ↓
SQLite Database
     ↓
Dashboard + LLM
```

### Why This is Perfect for EdgeSight
1. **Real simulation** of IoT environments
2. **Constraint testing** (lossy networks, slow sensors)
3. **Hardware scenario** (edge computing, offline-capable)
4. **LLM use case** - "Predict failure from noisy sensor data"

### 4.1 MQTT Integration Architecture

```go
// New file: internal/clients/mqtt_sensor.go
package clients

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTSensorClient struct {
	broker   string
	clientID string
	topics   []string
	values   map[string]float64
}

func NewMQTTSensorClient(broker string) *MQTTSensorClient {
	return &MQTTSensorClient{
		broker:   broker,
		clientID: "edgesight-collector",
		topics: []string{
			"sensors/temperature",
			"sensors/humidity",
			"sensors/pressure",
			"sensors/pm25",
			"sensors/power",
		},
		values: make(map[string]float64),
	}
}

func (c *MQTTSensorClient) Connect() error {
	opts := mqtt.NewClientOptions().
		AddBroker(c.broker).
		SetClientID(c.clientID)
	
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	
	// Subscribe to all topics
	for _, topic := range c.topics {
		client.Subscribe(topic, 1, c.messageHandler)
	}
	
	return nil
}

func (c *MQTTSensorClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	// Parse sensor value from topic/payload
	topic := msg.Topic()
	value := parseValue(msg.Payload())
	c.values[topic] = value
}

func (c *MQTTSensorClient) GetValues() map[string]float64 {
	return c.values
}
```

### 4.2 Sensor Simulator

**Separate tool to generate synthetic data with noise:**

```go
// New file: cmd/sensor-simulator/main.go
package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"math"
	"math/rand"
	"time"
)

type SensorSimulator struct {
	client mqtt.Client
	config SimConfig
}

type SimConfig struct {
	Temperature float64  // Base: 72F
	Humidity    float64  // Base: 45%
	Pressure    float64  // Base: 1013 hPa
	PM25        float64  // Base: 35 µg/m³
	Power       float64  // Base: 1000W
	
	NoiseLevel  float64  // 0.05 = ±5% noise
	Drift       float64  // 0.01 = 1% drift per hour
	DropRate    float64  // 0.1 = 10% packet loss
}

func (s *SensorSimulator) SimulateWithNoise() {
	ticker := time.NewTicker(5 * time.Second)
	
	for range ticker.C {
		// Apply drift over time
		s.config.Temperature += s.config.Temperature * s.config.Drift / 3600
		
		// Add noise to each reading
		temp := s.config.Temperature + randomNoise(s.config.Temperature, s.config.NoiseLevel)
		humidity := s.config.Humidity + randomNoise(s.config.Humidity, s.config.NoiseLevel)
		
		// Simulate packet loss
		if rand.Float64() > s.config.DropRate {
			s.publish("sensors/temperature", temp)
			s.publish("sensors/humidity", humidity)
		}
		// If packet lost, no publish (backend handles timeout)
	}
}

func randomNoise(base float64, noiseLevel float64) float64 {
	return base * noiseLevel * (rand.Float64()*2 - 1) // ±noiseLevel
}
```

### 4.3 Hardware Constraint Scenarios

**You could simulate different edge environments:**

```go
// Scenario 1: Stable data center
SimConfig{
	NoiseLevel: 0.01,   // ±1% noise
	DropRate:   0.001,  // 0.1% packet loss
	Drift:      0.0001, // minimal drift
}

// Scenario 2: Remote sensor (poor connectivity)
SimConfig{
	NoiseLevel: 0.15,   // ±15% noise
	DropRate:   0.3,    // 30% packet loss
	Drift:      0.05,   // significant drift
}

// Scenario 3: Industrial plant
SimConfig{
	NoiseLevel: 0.08,   // ±8% noise
	DropRate:   0.05,   // 5% packet loss
	Drift:      0.02,   // small drift
}
```

### 4.4 Integration with Canonicalizer

Add MQTT data to the snapshot:

```go
// In canonicalizer.go
type MQTTData struct {
	Temperature float64
	Humidity    float64
	Pressure    float64
	PM25        float64
	PowerUsage  float64
}

func BuildSnapshot(
	// ... existing params ...
	mqttData *clients.MQTTData,
) *models.Snapshot {
	// ... existing code ...
	
	// Add MQTT sensor data
	if mqttData != nil {
		snap.Environment.PM25 = int(mqttData.PM25)
		snap.Environment.Temperature = int(mqttData.Temperature)
		snap.Mobility.PowerUsage = int(mqttData.PowerUsage) // or new field
	}
	
	return snap
}
```

---

## 5. .NET SDK Option - Recommendation

You mentioned downloading .NET SDK. Here's the plan:

### Option A: Keep Go Backend, Add C# Frontend
- ✅ Go backend handles data ingestion (proven, optimized)
- ✅ C# Windows Forms/WPF for native Windows UI
- ✅ Cleaner separation of concerns
- ✅ Better for constraint simulation UI (sliders for noise, etc.)

**Build:** `go-ingest` backend + `csharp-ui` frontend + `sensor-simulator` tool

### Option B: Full .NET Stack
- ❌ Rewrite all Go clients in C# (unnecessary work)
- ❌ Different data layer patterns
- ❌ Less optimal for server-side ingestion

### Recommendation
**Stick with Option A:** Keep Go backend, test .NET SDK for an **interactive simulator UI**:

```csharp
// WPF form for sensor simulator
public class SensorSimulatorUI
{
    Slider NoiseSlider;      // 0-20% noise
    Slider DriftSlider;      // 0-5% drift
    Slider PacketLossSlider; // 0-50% loss
    
    void OnSimulateClick()
    {
        var config = new SimConfig
        {
            NoiseLevel = NoiseSlider.Value / 100,
            Drift = DriftSlider.Value / 100,
            DropRate = PacketLossSlider.Value / 100,
        };
        
        // Publish to MQTT broker
        simulator.PublishWithConfig(config);
    }
}
```

This gives you **interactive constraint testing** without rewriting the backend.

---

## 6. Implementation Roadmap - Phase 2

### Phase 2a: MQTT Integration (Week 1)
```
□ Add paho-mqtt Go dependency
□ Create MQTTSensorClient
□ Create sensor-simulator tool
□ Test with local MQTT broker (Mosquitto)
□ Integrate MQTT data into canonicalizer
□ Update API to expose MQTT metrics
□ Dashboard displays real/simulated sensors
```

### Phase 2b: Vector Embeddings (Week 2)
```
□ Choose embedding model (all-MiniLM-L6-v2)
□ Add sqlite-vec extension to schema
□ Create semantic summary generator
□ Create embedder service
□ Modify ingest to generate embeddings
□ Test vector search
□ Add /api/v1/search endpoint
```

### Phase 2c: LLM Integration (Week 2-3)
```
□ Set up go-llama.cpp bindings
□ Create LLM engine service
□ Load Qwen 2.5 7B model
□ Create query handler with context injection
□ Add /api/v1/query endpoint
□ Test end-to-end: query → embeddings → LLM → answer
```

### Phase 2d: UI Enhancements (Week 3)
```
□ Add query box to dashboard
□ Add response display area
□ Add source attribution
□ Add simulation controls (if using C#)
□ Test .NET SDK for Windows Forms UI
```

---

## 7. Key Technical Decisions

| Decision | Choice | Reasoning |
|----------|--------|-----------|
| **LLM Model** | Qwen 2.5 7B | Good balance: 7B = runnable locally, strong instruction following |
| **LLM Loading** | go-llama.cpp | Direct GGUF loading, no Ollama needed, lighter stack |
| **Embeddings** | all-MiniLM-L6-v2 | 384-dim, small, fast, good for domain understanding |
| **Vector DB** | SQLite-vec | Integrated with existing DB, no external service |
| **MQTT Broker** | Mosquitto (local) | Lightweight, perfect for simulation |
| **Sensor Sim** | Go CLI tool | Consistency with backend, no extra runtime |
| **Frontend UI** | Vanilla JS (Phase 2) | Keep consistency, add C# later if desired |
| **Constraint Sim** | Configurable noise/drift | Realistic IoT scenarios, edge computing testing |

---

## 8. Additional Thoughts & Suggestions

### 8.1 What You Should Do Now (Before Phase 2 Code)
1. **Update Movebank** - Add credentials to `.env` for completeness
2. **Optional: Fix CDC** - Use the JSON endpoint (Option A) for robustness
3. **Test .NET SDK** - Confirm it installs/works, plan UI layer
4. **Verify MQTT** - Download Mosquitto, test local broker

### 8.2 LLM Use Cases for EdgeSight

Given your hardware/constraints focus:

**1. Anomaly Detection**
```
User: "Are any sensors malfunctioning?"
LLM: Analyzes last 24h of data, looks for impossible patterns
Answer: "Temperature sensor seems stuck at 72°F for 6 hours - possible malfunction"
```

**2. Predictive Alerts**
```
User: "Will power usage spike tomorrow?"
LLM: Considers historical patterns, weather forecast
Answer: "Based on temperature forecast (low 35°F), expect 40% power increase"
```

**3. Hardware Recommendations**
```
User: "What hardware should I upgrade?"
LLM: Analyzes constraint patterns, packet loss
Answer: "Network bandwidth is limiting factor (30% drops). Consider fiber upgrade"
```

**4. Constraint Impact Analysis**
```
User: "How do dropouts affect our monitoring?"
LLM: Simulates scenarios with current noise/drift settings
Answer: "At 30% packet loss, you'll miss critical readings. Need buffering"
```

### 8.3 Safety Considerations for Local LLM

When running Qwen locally:
- ✅ No data leaves your machine
- ✅ No API rate limits
- ✅ Runs offline
- ⚠️ Needs GPU (or CPU with patience) - Qwen 7B is heavy
- ⚠️ First inference is slow (model loading)

**Optimize for speed:**
```go
// Keep model loaded in memory
var globalLLMEngine *llm.Engine

func init() {
    globalLLMEngine, _ = llm.NewLLMEngine("models/qwen2.5-7b.gguf")
}

// Reuse for queries
func (s *APIServer) HandleQuery(...) {
    answer, _ := globalLLMEngine.Query(prompt)
}
```

### 8.4 Testing Strategy

Before deploying Phase 2:

```bash
# 1. Test Qwen inference locally
go run cmd/test-llm/main.go

# 2. Test embeddings
go run cmd/test-embeddings/main.go

# 3. Test vector search
go run cmd/test-vector-search/main.go

# 4. Integration test
go run cmd/test-integration/main.go
```

### 8.5 Alternative to Consider: Smaller LLM

If Qwen 7B is too heavy, alternatives:
- **Phi-2** (2.7B, Microsoft, very capable)
- **TinyLlama** (1.1B, fast, lower VRAM)
- **Mistral 7B** (slightly larger but very smart)

For constraint/hardware simulation, 7B is probably right choice.

---

## Summary

| Aspect | Status | Next Step |
|--------|--------|-----------|
| **API Keys** | 90% done | Add Movebank credentials |
| **CDC Issue** | Identified | Use JSON endpoint instead of POST |
| **LLM Setup** | Ready to go | Load Qwen 7B with go-llama.cpp |
| **Vector DB** | Designed | Add sqlite-vec extension |
| **MQTT Integration** | Designed | Implement broker + simulator |
| **Constraint Testing** | Planned | Create noise/drift scenarios |
| **Phase 2 UI** | Deferred | Use C# for simulator UI, vanilla JS for queries |

---

**Next Action:** Confirm you want to proceed with:
1. Movebank credential setup
2. MQTT + sensor simulator integration first
3. Then vector embeddings + LLM

---
