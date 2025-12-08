package store

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// SnapshotEmbedding represents an embedding linked to a snapshot.
type SnapshotEmbedding struct {
	ID         int64
	SnapshotTS string
	Location   string
	Summary    string
	Embedding  []float64
	CreatedAt  time.Time
}

// SearchResult includes similarity score for a snapshot embedding.
type SearchResult struct {
	SnapshotEmbedding
	Score float64
}

// InsertEmbedding stores an embedding for a snapshot.
func (s *SQLiteStore) InsertEmbedding(e SnapshotEmbedding) error {
	blob, err := json.Marshal(e.Embedding)
	if err != nil {
		return fmt.Errorf("marshal embedding: %w", err)
	}
	_, err = s.DB.Exec(`INSERT INTO snapshot_embeddings (snapshot_ts, location, summary, embedding, created_at) VALUES (?, ?, ?, ?, ?)`,
		e.SnapshotTS, e.Location, e.Summary, string(blob), e.CreatedAt.Format(time.RFC3339))
	return err
}

// GetEmbeddingsByLocation fetches embeddings for a location (optionally limit recent).
func (s *SQLiteStore) GetEmbeddingsByLocation(location string, limit int) ([]SnapshotEmbedding, error) {
	q := `SELECT id, snapshot_ts, location, summary, embedding, created_at FROM snapshot_embeddings WHERE location = ? ORDER BY created_at DESC`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.DB.Query(q, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SnapshotEmbedding
	for rows.Next() {
		var rec SnapshotEmbedding
		var embText string
		var created string
		if err := rows.Scan(&rec.ID, &rec.SnapshotTS, &rec.Location, &rec.Summary, &embText, &created); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(embText), &rec.Embedding); err != nil {
			return nil, err
		}
		if ts, err := time.Parse(time.RFC3339, created); err == nil {
			rec.CreatedAt = ts
		}
		out = append(out, rec)
	}
	return out, nil
}

// SearchEmbeddings naive cosine similarity search in Go (acceptable for small N).
func (s *SQLiteStore) SearchEmbeddings(location string, queryVec []float64, topK int) ([]SearchResult, error) {
	recs, err := s.GetEmbeddingsByLocation(location, 0)
	if err != nil {
		return nil, err
	}
	type scored struct {
		rec   SnapshotEmbedding
		score float64
	}
	var scoredList []scored
	for _, r := range recs {
		if len(r.Embedding) == 0 || len(r.Embedding) != len(queryVec) {
			continue
		}
		scoredList = append(scoredList, scored{rec: r, score: cosine(queryVec, r.Embedding)})
	}
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].score > scoredList[j].score })
	if topK > 0 && len(scoredList) > topK {
		scoredList = scoredList[:topK]
	}
	out := make([]SearchResult, len(scoredList))
	for i, s := range scoredList {
		out[i] = SearchResult{SnapshotEmbedding: s.rec, Score: s.score}
	}
	return out, nil
}

func cosine(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
