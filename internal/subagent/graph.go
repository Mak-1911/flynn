// Package subagent provides the GraphAgent for knowledge graph operations.
package subagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flynn-ai/flynn/internal/memory"
)

// GraphAgent handles knowledge graph operations.
type GraphAgent struct {
	store *memory.GraphStore
}

// NewGraphAgent creates a new graph subagent.
func NewGraphAgent(store *memory.GraphStore) *GraphAgent {
	return &GraphAgent{store: store}
}

// Name returns the subagent name.
func (g *GraphAgent) Name() string {
	return "graph"
}

// Description returns the subagent description.
func (g *GraphAgent) Description() string {
	return "Handles knowledge graph operations: ingest, search, link, stats"
}

// Capabilities returns the list of supported actions.
func (g *GraphAgent) Capabilities() []string {
	return []string{
		"ingest_file",   // Ingest a file into the document store
		"ingest_text",   // Ingest raw text content
		"entity_upsert", // Create or update an entity
		"link",          // Link two entities
		"search",        // Search entities by name
		"relations",     // Get relations for an entity
		"stats",         // Graph stats
	}
}

// ValidateAction checks if an action is supported.
func (g *GraphAgent) ValidateAction(action string) bool {
	for _, cap := range g.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// Execute executes a graph operation.
func (g *GraphAgent) Execute(ctx context.Context, step *PlanStep) (*Result, error) {
	startTime := time.Now()

	if !g.ValidateAction(step.Action) {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", step.Action),
		}, nil
	}

	if g.store == nil {
		return &Result{Success: false, Error: "graph store not initialized"}, nil
	}

	tenantID := "default"
	if t, ok := step.Input["tenant_id"].(string); ok && t != "" {
		tenantID = t
	}

	var result any
	var err error

	switch step.Action {
	case "ingest_file":
		path, ok := step.Input["path"].(string)
		if !ok || path == "" {
			return &Result{Success: false, Error: "path parameter required"}, nil
		}
		result, err = g.ingestFile(ctx, tenantID, path)
	case "ingest_text":
		content, ok := step.Input["content"].(string)
		if !ok || content == "" {
			return &Result{Success: false, Error: "content parameter required"}, nil
		}
		path, _ := step.Input["path"].(string)
		if path == "" {
			path = fmt.Sprintf("memory://note-%d", time.Now().UnixNano())
		}
		title, _ := step.Input["title"].(string)
		result, err = g.ingestText(ctx, tenantID, path, title, content)
	case "entity_upsert":
		name, ok := step.Input["name"].(string)
		if !ok || name == "" {
			return &Result{Success: false, Error: "name parameter required"}, nil
		}
		entityType, ok := step.Input["type"].(string)
		if !ok || entityType == "" {
			return &Result{Success: false, Error: "type parameter required"}, nil
		}
		desc, _ := step.Input["description"].(string)
		result, err = g.upsertEntity(ctx, tenantID, name, entityType, desc)
	case "link":
		sourceName, ok := step.Input["source_name"].(string)
		if !ok || sourceName == "" {
			return &Result{Success: false, Error: "source_name parameter required"}, nil
		}
		sourceType, ok := step.Input["source_type"].(string)
		if !ok || sourceType == "" {
			return &Result{Success: false, Error: "source_type parameter required"}, nil
		}
		targetName, ok := step.Input["target_name"].(string)
		if !ok || targetName == "" {
			return &Result{Success: false, Error: "target_name parameter required"}, nil
		}
		targetType, ok := step.Input["target_type"].(string)
		if !ok || targetType == "" {
			return &Result{Success: false, Error: "target_type parameter required"}, nil
		}
		relationType, ok := step.Input["relation_type"].(string)
		if !ok || relationType == "" {
			return &Result{Success: false, Error: "relation_type parameter required"}, nil
		}
		result, err = g.linkEntities(ctx, tenantID, sourceName, sourceType, targetName, targetType, relationType)
	case "search":
		query, ok := step.Input["query"].(string)
		if !ok || query == "" {
			return &Result{Success: false, Error: "query parameter required"}, nil
		}
		result, err = g.searchEntities(ctx, tenantID, query)
	case "relations":
		entityID, _ := step.Input["entity_id"].(string)
		if entityID == "" {
			name, _ := step.Input["name"].(string)
			entityType, _ := step.Input["type"].(string)
			if name == "" || entityType == "" {
				return &Result{Success: false, Error: "entity_id or (name and type) required"}, nil
			}
			entity, findErr := g.store.FindEntityByName(ctx, tenantID, name, entityType)
			if findErr != nil {
				return &Result{Success: false, Error: findErr.Error()}, nil
			}
			if entity == nil {
				return &Result{Success: false, Error: "entity not found"}, nil
			}
			entityID = entity.ID
		}
		result, err = g.getRelations(ctx, tenantID, entityID)
	case "stats":
		result, err = g.store.Stats(ctx, tenantID)
	default:
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("action not implemented: %s", step.Action),
		}, nil
	}

	if err != nil {
		return &Result{
			Success:    false,
			Error:      err.Error(),
			DurationMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &Result{
		Success:    true,
		Data:       result,
		DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// ============================================================
// Action Implementations
// ============================================================

func (g *GraphAgent) ingestFile(ctx context.Context, tenantID, path string) (any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	fileType := strings.TrimPrefix(filepath.Ext(absPath), ".")
	title := filepath.Base(absPath)

	return g.ingestText(ctx, tenantID, absPath, title, string(content))
}

func (g *GraphAgent) ingestText(ctx context.Context, tenantID, path, title, content string) (any, error) {
	preview := content
	if len(preview) > 500 {
		preview = preview[:500]
	}

	doc := &memory.Document{
		Path:           path,
		Title:          title,
		ContentPreview: preview,
		FileType:       strings.TrimPrefix(filepath.Ext(path), "."),
		SizeBytes:      len(content),
		Language:       guessLanguage(path),
		ChunkCount:     0,
	}

	chunks := chunkText(content, 1000)
	doc.ChunkCount = len(chunks)

	saved, err := g.store.UpsertDocument(ctx, tenantID, doc)
	if err != nil {
		return nil, err
	}

	var chunkRows []memory.DocumentChunk
	for i, chunk := range chunks {
		chunkRows = append(chunkRows, memory.DocumentChunk{
			DocumentID: saved.ID,
			Index:      i,
			Content:    chunk,
		})
	}

	if err := g.store.ReplaceDocumentChunks(ctx, tenantID, saved.ID, chunkRows); err != nil {
		return nil, err
	}

	return map[string]any{
		"document_id": saved.ID,
		"path":        saved.Path,
		"chunks":      len(chunkRows),
	}, nil
}

func (g *GraphAgent) upsertEntity(ctx context.Context, tenantID, name, entityType, description string) (any, error) {
	entity, err := g.store.UpsertEntity(ctx, tenantID, &memory.Entity{
		Name:        name,
		EntityType:  entityType,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (g *GraphAgent) linkEntities(ctx context.Context, tenantID, sourceName, sourceType, targetName, targetType, relationType string) (any, error) {
	source, err := g.store.UpsertEntity(ctx, tenantID, &memory.Entity{
		Name:       sourceName,
		EntityType: sourceType,
	})
	if err != nil {
		return nil, err
	}
	target, err := g.store.UpsertEntity(ctx, tenantID, &memory.Entity{
		Name:       targetName,
		EntityType: targetType,
	})
	if err != nil {
		return nil, err
	}

	rel, err := g.store.CreateRelation(ctx, tenantID, &memory.Relation{
		SourceID:     source.ID,
		TargetID:     target.ID,
		RelationType: relationType,
	})
	if err != nil {
		return nil, err
	}

	return rel, nil
}

func (g *GraphAgent) searchEntities(ctx context.Context, tenantID, query string) (any, error) {
	return g.store.SearchEntities(ctx, tenantID, query, 20)
}

func (g *GraphAgent) getRelations(ctx context.Context, tenantID, entityID string) (any, error) {
	return g.store.GetRelations(ctx, tenantID, entityID, 50)
}

// ============================================================
// Helpers
// ============================================================

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

func guessLanguage(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		return "text"
	}
	switch ext {
	case "go", "js", "ts", "tsx", "jsx", "py", "rs", "java", "kt", "cs", "rb", "php":
		return ext
	case "md", "txt":
		return "text"
	default:
		return ext
	}
}
