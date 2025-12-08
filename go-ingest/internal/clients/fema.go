package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// FEMAClient reads FEMA disaster summaries from a static JSON file.
type FEMAClient struct {
	dataPath string
}

// FEMASummary aggregates key disaster metrics for a state.
type FEMASummary struct {
	ActiveDisasters  int
	TopIncidentType  string
	Severity         int
	AffectedCounties int
}

type femaPayload struct {
	DisasterDeclarationsSummaries []femaRecord `json:"DisasterDeclarationsSummaries"`
}

type femaRecord struct {
	State             string  `json:"state"`
	IncidentType      string  `json:"incidentType"`
	DeclarationType   string  `json:"declarationType"`
	IncidentBeginDate string  `json:"incidentBeginDate"`
	DisasterCloseout  *string `json:"disasterCloseoutDate"`
	FIPSCountyCode    string  `json:"fipsCountyCode"`
}

// NewFEMAClient creates a client pointing at a FEMA JSON export; defaults to the repo root file when path is empty.
func NewFEMAClient(path string) *FEMAClient {
	if path == "" {
		path = "DisasterDeclarationsSummaries.json"
	}
	return &FEMAClient{dataPath: path}
}

// GetStateSummary returns a lightweight summary for the requested state.
// lookbackDays scopes how far back we consider events; default is 180 days when <= 0.
func (c *FEMAClient) GetStateSummary(state string, lookbackDays int) (*FEMASummary, error) {
	state = strings.ToUpper(strings.TrimSpace(state))
	if state == "" {
		return nil, fmt.Errorf("state code required")
	}
	if lookbackDays <= 0 {
		lookbackDays = 180
	}

	f, err := os.Open(c.dataPath)
	if err != nil {
		return nil, fmt.Errorf("open FEMA file: %w", err)
	}
	defer f.Close()

	var payload femaPayload
	if err := json.NewDecoder(f).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode FEMA payload: %w", err)
	}

	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -lookbackDays)
	typeCounts := make(map[string]int)
	counties := make(map[string]struct{})
	maxSeverity := 0
	active := 0

	for _, rec := range payload.DisasterDeclarationsSummaries {
		if strings.ToUpper(rec.State) != state {
			continue
		}

		begin := parseFEMATime(rec.IncidentBeginDate)
		closeout := parseFEMATimePtr(rec.DisasterCloseout)

		if begin.IsZero() {
			continue
		}

		// Treat events as relevant if recent or still open.
		isRecent := begin.After(cutoff)
		isOpen := closeout.IsZero() || closeout.After(now)
		if !isRecent && !isOpen {
			continue
		}

		active++
		typeCounts[rec.IncidentType]++

		if rec.FIPSCountyCode != "" && rec.FIPSCountyCode != "000" {
			counties[rec.FIPSCountyCode] = struct{}{}
		}

		sev := severityFromDeclaration(rec.DeclarationType)
		if sev > maxSeverity {
			maxSeverity = sev
		}
	}

	return &FEMASummary{
		ActiveDisasters:  active,
		TopIncidentType:  selectTopIncident(typeCounts),
		Severity:         maxSeverity,
		AffectedCounties: len(counties),
	}, nil
}

func parseFEMATime(val string) time.Time {
	if val == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseFEMATimePtr(val *string) time.Time {
	if val == nil {
		return time.Time{}
	}
	return parseFEMATime(*val)
}

// severityFromDeclaration maps FEMA declaration types into a coarse 1-5 scale.
func severityFromDeclaration(declType string) int {
	switch strings.ToUpper(declType) {
	case "DR": // Major Disaster Declaration
		return 5
	case "EM": // Emergency Declaration
		return 4
	case "FM": // Fire Management Assistance
		return 3
	case "FS": // Fire Suppression (legacy)
		return 2
	default:
		return 1
	}
}

func selectTopIncident(counts map[string]int) string {
	var top string
	var max int
	for k, v := range counts {
		if v > max {
			max = v
			top = k
		}
	}
	return top
}
