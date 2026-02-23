// Package graph provides ingestion pipeline for the knowledge graph.
package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/memory"
)

// Source identifies where the content came from.
type Source struct {
	Type string
	Ref  string
}

// Ingestor handles document storage and entity/relation extraction.
type Ingestor struct {
	Store         *memory.GraphStore
	Extractor     Extractor
	MaxEntities   int
	MaxRelations  int
	MaxChunkBytes int
}

// NewIngestor creates a new graph ingestor.
func NewIngestor(store *memory.GraphStore, extractor Extractor) *Ingestor {
	return &Ingestor{
		Store:         store,
		Extractor:     extractor,
		MaxEntities:   20,
		MaxRelations:  20,
		MaxChunkBytes: 2000,
	}
}

// IngestText ingests raw text into document storage and extracts graph facts.
func (i *Ingestor) IngestText(ctx context.Context, tenantID string, source Source, title string, content string) (*IngestResult, error) {
	if i == nil || i.Store == nil {
		return nil, fmt.Errorf("ingestor not initialized")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is empty")
	}

	path := source.Ref
	if path == "" {
		path = fmt.Sprintf("text-%d", time.Now().UnixNano())
	}
	if source.Type != "" && !strings.Contains(path, "://") {
		path = fmt.Sprintf("%s://%s", source.Type, path)
	}

	doc := &memory.Document{
		Path:           path,
		Title:          title,
		ContentPreview: preview(content, 500),
		FileType:       "text",
		SizeBytes:      len(content),
		Language:       "text",
		ChunkCount:     0,
	}

	chunkSize := i.MaxChunkBytes
	if chunkSize <= 0 {
		chunkSize = 2000
	}
	chunks := chunkText(content, chunkSize)
	doc.ChunkCount = len(chunks)

	saved, err := i.Store.UpsertDocument(ctx, tenantID, doc)
	if err != nil {
		return nil, err
	}

	var chunkRows []memory.DocumentChunk
	for idx, chunk := range chunks {
		chunkRows = append(chunkRows, memory.DocumentChunk{
			DocumentID: saved.ID,
			Index:      idx,
			Content:    chunk,
		})
	}

	if err := i.Store.ReplaceDocumentChunks(ctx, tenantID, saved.ID, chunkRows); err != nil {
		return nil, err
	}

	entities, relations, err := i.extract(ctx, content)
	if err != nil {
		return nil, err
	}

	entityCount, relationCount, err := i.persistFacts(ctx, tenantID, entities, relations)
	if err != nil {
		return nil, err
	}

	return &IngestResult{
		DocumentID: saved.ID,
		Chunks:     len(chunkRows),
		Entities:   entityCount,
		Relations:  relationCount,
	}, nil
}

// IngestResult summarizes ingestion output.
type IngestResult struct {
	DocumentID string
	Chunks     int
	Entities   int
	Relations  int
}

func (i *Ingestor) extract(ctx context.Context, content string) ([]ExtractedEntity, []ExtractedRelation, error) {
	if i.Extractor == nil {
		return nil, nil, nil
	}
	entities, relations, err := i.Extractor.Extract(ctx, content)
	if err != nil {
		return nil, nil, err
	}

	if i.MaxEntities > 0 && len(entities) > i.MaxEntities {
		entities = entities[:i.MaxEntities]
	}
	if i.MaxRelations > 0 && len(relations) > i.MaxRelations {
		relations = relations[:i.MaxRelations]
	}

	return entities, relations, nil
}

func (i *Ingestor) persistFacts(ctx context.Context, tenantID string, entities []ExtractedEntity, relations []ExtractedRelation) (int, int, error) {
	entityMap := make(map[string]*memory.Entity)

	for _, e := range entities {
		if strings.TrimSpace(e.Name) == "" {
			continue
		}
		entity, err := i.Store.UpsertEntity(ctx, tenantID, &memory.Entity{
			Name:        e.Name,
			EntityType:  e.Type,
			Description: e.Description,
		})
		if err != nil {
			return 0, 0, err
		}
		key := fmt.Sprintf("%s|%s", strings.ToLower(entity.Name), entity.EntityType)
		entityMap[key] = entity
	}

	relationCount := 0
	for _, r := range relations {
		sourceKey := fmt.Sprintf("%s|%s", strings.ToLower(r.SourceName), r.SourceType)
		targetKey := fmt.Sprintf("%s|%s", strings.ToLower(r.TargetName), r.TargetType)

		source := entityMap[sourceKey]
		target := entityMap[targetKey]
		if source == nil || target == nil {
			continue
		}

		if _, err := i.Store.CreateRelation(ctx, tenantID, &memory.Relation{
			SourceID:     source.ID,
			TargetID:     target.ID,
			RelationType: r.Relation,
		}); err != nil {
			return 0, relationCount, err
		}
		relationCount++
	}

	return len(entityMap), relationCount, nil
}

func preview(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max]
}

func chunkText(text string, maxLen int) []string {
	if maxLen <= 0 {
		return []string{text}
	}
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += maxLen {
		end := i + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	if len(chunks) == 0 {
		chunks = []string{""}
	}
	return chunks
}
