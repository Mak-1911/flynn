// Package memory provides knowledge graph storage helpers.
package memory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// GraphStore provides CRUD helpers for the knowledge graph tables.
type GraphStore struct {
	db *sql.DB
}

// NewGraphStore creates a graph store backed by the team database.
func NewGraphStore(db *sql.DB) *GraphStore {
	return &GraphStore{db: db}
}

// Entity represents a node in the knowledge graph.
type Entity struct {
	ID           string
	TenantID     string
	Name         string
	EntityType   string
	Description  string
	MetadataJSON string
	EmbeddingID  string
	Importance   float64
	CreatedAt    int64
	UpdatedAt    int64
}

// Relation represents an edge in the knowledge graph.
type Relation struct {
	ID           string
	TenantID     string
	SourceID     string
	TargetID     string
	RelationType string
	MetadataJSON string
	Confidence   float64
	CreatedAt    int64
	UpdatedAt    int64
}

// Document represents an indexed document.
type Document struct {
	ID             string
	TenantID       string
	Path           string
	Title          string
	ContentPreview string
	FileType       string
	SizeBytes      int
	Language       string
	IndexedAt      int64
	UpdatedAt      int64
	ChunkCount     int
	MetadataJSON   string
}

// DocumentChunk represents a document chunk.
type DocumentChunk struct {
	DocumentID   string
	Index        int
	Content      string
	MetadataJSON string
}

// UpsertEntity inserts or updates an entity by (tenant_id, name, entity_type).
func (g *GraphStore) UpsertEntity(ctx context.Context, tenantID string, entity *Entity) (*Entity, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}
	if entity == nil {
		return nil, fmt.Errorf("entity is required")
	}
	if entity.Name == "" || entity.EntityType == "" {
		return nil, fmt.Errorf("entity name and type are required")
	}

	now := time.Now().Unix()
	if entity.ID == "" {
		entity.ID = fmt.Sprintf("ent-%d", time.Now().UnixNano())
	}

	entity.TenantID = tenantID
	entity.UpdatedAt = now
	if entity.CreatedAt == 0 {
		entity.CreatedAt = now
	}

	_, err := g.db.ExecContext(ctx, `
		INSERT INTO team_entities (id, tenant_id, name, entity_type, description, metadata_json, embedding_id, importance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tenant_id, name, entity_type)
		DO UPDATE SET
			description = excluded.description,
			metadata_json = excluded.metadata_json,
			embedding_id = excluded.embedding_id,
			importance = excluded.importance,
			updated_at = excluded.updated_at
	`, entity.ID, tenantID, entity.Name, entity.EntityType, entity.Description, entity.MetadataJSON, entity.EmbeddingID, entity.Importance, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// CreateRelation inserts a relation edge between two entities.
func (g *GraphStore) CreateRelation(ctx context.Context, tenantID string, relation *Relation) (*Relation, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}
	if relation == nil {
		return nil, fmt.Errorf("relation is required")
	}
	if relation.SourceID == "" || relation.TargetID == "" || relation.RelationType == "" {
		return nil, fmt.Errorf("source_id, target_id, and relation_type are required")
	}

	now := time.Now().Unix()
	if relation.ID == "" {
		relation.ID = fmt.Sprintf("rel-%d", time.Now().UnixNano())
	}

	relation.TenantID = tenantID
	relation.CreatedAt = now
	relation.UpdatedAt = now
	if relation.Confidence == 0 {
		relation.Confidence = 1.0
	}

	_, err := g.db.ExecContext(ctx, `
		INSERT INTO team_relations (id, tenant_id, source_id, target_id, relation_type, metadata_json, confidence, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, relation.ID, tenantID, relation.SourceID, relation.TargetID, relation.RelationType, relation.MetadataJSON, relation.Confidence, relation.CreatedAt, relation.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return relation, nil
}

// FindEntityByName returns a single entity by exact name and type.
func (g *GraphStore) FindEntityByName(ctx context.Context, tenantID, name, entityType string) (*Entity, error) {
	var e Entity
	err := g.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, entity_type, description, metadata_json, embedding_id, importance, created_at, updated_at
		FROM team_entities
		WHERE tenant_id = ? AND name = ? AND entity_type = ?
		LIMIT 1
	`, tenantID, name, entityType).Scan(
		&e.ID, &e.TenantID, &e.Name, &e.EntityType, &e.Description, &e.MetadataJSON,
		&e.EmbeddingID, &e.Importance, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetEntityByID returns a single entity by ID.
func (g *GraphStore) GetEntityByID(ctx context.Context, tenantID, id string) (*Entity, error) {
	var e Entity
	err := g.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, entity_type, description, metadata_json, embedding_id, importance, created_at, updated_at
		FROM team_entities
		WHERE tenant_id = ? AND id = ?
		LIMIT 1
	`, tenantID, id).Scan(
		&e.ID, &e.TenantID, &e.Name, &e.EntityType, &e.Description, &e.MetadataJSON,
		&e.EmbeddingID, &e.Importance, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// SearchEntities performs a simple LIKE search on entity names.
func (g *GraphStore) SearchEntities(ctx context.Context, tenantID, query string, limit int) ([]*Entity, error) {
	if limit <= 0 {
		limit = 20
	}
	like := "%" + strings.TrimSpace(query) + "%"

	rows, err := g.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, entity_type, description, metadata_json, embedding_id, importance, created_at, updated_at
		FROM team_entities
		WHERE tenant_id = ? AND name LIKE ?
		ORDER BY importance DESC, updated_at DESC
		LIMIT ?
	`, tenantID, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.Name, &e.EntityType, &e.Description, &e.MetadataJSON,
			&e.EmbeddingID, &e.Importance, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// GetRelations returns relations for an entity (as source or target).
func (g *GraphStore) GetRelations(ctx context.Context, tenantID, entityID string, limit int) ([]*Relation, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := g.db.QueryContext(ctx, `
		SELECT id, tenant_id, source_id, target_id, relation_type, metadata_json, confidence, created_at, updated_at
		FROM team_relations
		WHERE tenant_id = ? AND (source_id = ? OR target_id = ?)
		ORDER BY updated_at DESC
		LIMIT ?
	`, tenantID, entityID, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Relation
	for rows.Next() {
		var r Relation
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.SourceID, &r.TargetID, &r.RelationType, &r.MetadataJSON,
			&r.Confidence, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// UpsertDocument inserts or updates a document by (tenant_id, path).
func (g *GraphStore) UpsertDocument(ctx context.Context, tenantID string, doc *Document) (*Document, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}
	if doc == nil {
		return nil, fmt.Errorf("document is required")
	}
	if strings.TrimSpace(doc.Path) == "" {
		return nil, fmt.Errorf("document path is required")
	}

	now := time.Now().Unix()
	doc.TenantID = tenantID
	doc.UpdatedAt = now
	if doc.IndexedAt == 0 {
		doc.IndexedAt = now
	}

	existingID, err := g.getDocumentID(ctx, tenantID, doc.Path)
	if err != nil {
		return nil, err
	}
	if existingID != "" {
		doc.ID = existingID
	} else if doc.ID == "" {
		doc.ID = fmt.Sprintf("doc-%d", time.Now().UnixNano())
	}

	_, err = g.db.ExecContext(ctx, `
		INSERT INTO team_documents (id, tenant_id, path, title, content_preview, file_type, size_bytes, language, indexed_at, updated_at, chunk_count, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tenant_id, path)
		DO UPDATE SET
			title = excluded.title,
			content_preview = excluded.content_preview,
			file_type = excluded.file_type,
			size_bytes = excluded.size_bytes,
			language = excluded.language,
			indexed_at = excluded.indexed_at,
			updated_at = excluded.updated_at,
			chunk_count = excluded.chunk_count,
			metadata_json = excluded.metadata_json
	`, doc.ID, tenantID, doc.Path, doc.Title, doc.ContentPreview, doc.FileType, doc.SizeBytes, doc.Language, doc.IndexedAt, doc.UpdatedAt, doc.ChunkCount, doc.MetadataJSON)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// ReplaceDocumentChunks removes existing chunks and inserts new ones.
func (g *GraphStore) ReplaceDocumentChunks(ctx context.Context, tenantID, documentID string, chunks []DocumentChunk) error {
	if g == nil || g.db == nil {
		return fmt.Errorf("graph store not initialized")
	}
	if documentID == "" {
		return fmt.Errorf("documentID is required")
	}

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM team_doc_chunks WHERE document_id = ?`, documentID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO team_doc_chunks (id, tenant_id, document_id, chunk_index, content, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		chunkID := fmt.Sprintf("chunk-%d", time.Now().UnixNano())
		if _, err = stmt.ExecContext(ctx, chunkID, tenantID, documentID, chunk.Index, chunk.Content, chunk.MetadataJSON); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GraphStats provides summary counts for the graph tables.
type GraphStats struct {
	Entities  int
	Relations int
	Documents int
	Chunks    int
}

// Stats returns total counts for graph-related tables.
func (g *GraphStore) Stats(ctx context.Context, tenantID string) (*GraphStats, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}

	stats := &GraphStats{}
	if err := g.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_entities WHERE tenant_id = ?`, tenantID).Scan(&stats.Entities); err != nil {
		return nil, err
	}
	if err := g.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_relations WHERE tenant_id = ?`, tenantID).Scan(&stats.Relations); err != nil {
		return nil, err
	}
	if err := g.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_documents WHERE tenant_id = ?`, tenantID).Scan(&stats.Documents); err != nil {
		return nil, err
	}
	if err := g.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_doc_chunks WHERE tenant_id = ?`, tenantID).Scan(&stats.Chunks); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListEntities returns recent entities for a tenant.
func (g *GraphStore) ListEntities(ctx context.Context, tenantID string, limit int) ([]*Entity, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := g.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, entity_type, description, metadata_json, embedding_id, importance, created_at, updated_at
		FROM team_entities
		WHERE tenant_id = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.Name, &e.EntityType, &e.Description, &e.MetadataJSON,
			&e.EmbeddingID, &e.Importance, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// ListRelations returns recent relations for a tenant.
func (g *GraphStore) ListRelations(ctx context.Context, tenantID string, limit int) ([]*Relation, error) {
	if g == nil || g.db == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := g.db.QueryContext(ctx, `
		SELECT id, tenant_id, source_id, target_id, relation_type, metadata_json, confidence, created_at, updated_at
		FROM team_relations
		WHERE tenant_id = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Relation
	for rows.Next() {
		var r Relation
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.SourceID, &r.TargetID, &r.RelationType, &r.MetadataJSON,
			&r.Confidence, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

func (g *GraphStore) getDocumentID(ctx context.Context, tenantID, path string) (string, error) {
	var id string
	err := g.db.QueryRowContext(ctx, `
		SELECT id FROM team_documents WHERE tenant_id = ? AND path = ? LIMIT 1
	`, tenantID, path).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}
