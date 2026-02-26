// Package executor provides tool implementations for graph operations.
package executor

import (
	"context"
	"fmt"
	"time"
)

// GraphStats shows knowledge graph statistics.
type GraphStats struct {
	graph GraphService
}

func (t *GraphStats) Name() string { return "graph_stats" }

func (t *GraphStats) Description() string { return "Show knowledge graph statistics" }

func (t *GraphStats) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	stats, err := t.graph.Stats(ctx)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(stats), start), nil
}

// GraphSearch searches the knowledge graph.
type GraphSearch struct {
	graph GraphService
}

func (t *GraphSearch) Name() string { return "graph_search" }

func (t *GraphSearch) Description() string { return "Search the knowledge graph" }

func (t *GraphSearch) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	query, ok := input["query"].(string)
	if !ok || query == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("query is required")), start), nil
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	limit := 10
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	results, err := t.graph.Search(ctx, query, limit)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"query":   query,
		"results": results,
		"count":   len(results),
	}), start), nil
}

// GraphDump exports graph data.
type GraphDump struct {
	graph GraphService
}

func (t *GraphDump) Name() string { return "graph_dump" }

func (t *GraphDump) Description() string { return "Export graph data" }

func (t *GraphDump) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	format, _ := input["format"].(string)
	if format == "" {
		format = "json"
	}

	limit := 100
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	data, err := t.graph.Dump(ctx, format, limit)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"format": format,
		"data":   data,
	}), start), nil
}

// GraphQuery queries graph relationships.
type GraphQuery struct {
	graph GraphService
}

func (t *GraphQuery) Name() string { return "graph_query" }

func (t *GraphQuery) Description() string { return "Query graph relationships" }

func (t *GraphQuery) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	entity, ok := input["entity"].(string)
	if !ok || entity == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("entity is required")), start), nil
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	relations, err := t.graph.QueryRelations(ctx, entity)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"entity":     entity,
		"relations":  relations,
		"count":      len(relations),
	}), start), nil
}

// GraphAddEntity adds an entity to the graph.
type GraphAddEntity struct {
	graph GraphService
}

func (t *GraphAddEntity) Name() string { return "graph_add_entity" }

func (t *GraphAddEntity) Description() string { return "Add entity to graph" }

func (t *GraphAddEntity) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	entity, ok := input["entity"].(string)
	if !ok || entity == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("entity is required")), start), nil
	}

	entityType, _ := input["type"].(string)
	if entityType == "" {
		entityType = "unknown"
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	err := t.graph.AddEntity(ctx, entity, entityType, input["properties"])
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"added": true,
		"entity": entity,
		"type": entityType,
	}), start), nil
}

// GraphAddRelation adds a relation to the graph.
type GraphAddRelation struct {
	graph GraphService
}

func (t *GraphAddRelation) Name() string { return "graph_add_relation" }

func (t *GraphAddRelation) Description() string { return "Add relation to graph" }

func (t *GraphAddRelation) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	from, ok := input["from"].(string)
	if !ok || from == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("from is required")), start), nil
	}

	to, ok := input["to"].(string)
	if !ok || to == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("to is required")), start), nil
	}

	relation, ok := input["relation"].(string)
	if !ok || relation == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("relation is required")), start), nil
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	err := t.graph.AddRelation(ctx, from, to, relation)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"added": true,
		"from": from,
		"to": to,
		"relation": relation,
	}), start), nil
}

// GraphExport exports graph to file.
type GraphExport struct {
	graph GraphService
}

func (t *GraphExport) Name() string { return "graph_export" }

func (t *GraphExport) Description() string { return "Export graph to file" }

func (t *GraphExport) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok || path == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("path is required")), start), nil
	}

	format, _ := input["format"].(string)
	if format == "" {
		format = "json"
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	err := t.graph.Export(ctx, path, format)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"exported": true,
		"path": path,
		"format": format,
	}), start), nil
}

// GraphImport imports graph from file.
type GraphImport struct {
	graph GraphService
}

func (t *GraphImport) Name() string { return "graph_import" }

func (t *GraphImport) Description() string { return "Import graph from file" }

func (t *GraphImport) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	path, ok := input["path"].(string)
	if !ok || path == "" {
		return TimedResult(NewErrorResult(fmt.Errorf("path is required")), start), nil
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	count, err := t.graph.Import(ctx, path)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"imported": true,
		"path": path,
		"count": count,
	}), start), nil
}

// GraphClear clears all graph data.
type GraphClear struct {
	graph GraphService
}

func (t *GraphClear) Name() string { return "graph_clear" }

func (t *GraphClear) Description() string { return "Clear all graph data" }

func (t *GraphClear) Execute(ctx context.Context, input map[string]any) (*Result, error) {
	start := time.Now()

	confirm, ok := input["confirm"].(bool)
	if !ok || !confirm {
		return TimedResult(NewErrorResult(fmt.Errorf("confirm must be true to clear graph")), start), nil
	}

	if t.graph == nil {
		return TimedResult(NewErrorResult(fmt.Errorf("graph service not available")), start), nil
	}

	err := t.graph.Clear(ctx)
	if err != nil {
		return TimedResult(NewErrorResult(err), start), nil
	}

	return TimedResult(NewSuccessResult(map[string]any{
		"cleared": true,
	}), start), nil
}

// GraphService defines the interface for graph operations.
// This will be implemented by the actual graph service.
type GraphService interface {
	Stats(ctx context.Context) (map[string]any, error)
	Search(ctx context.Context, query string, limit int) ([]map[string]any, error)
	Dump(ctx context.Context, format string, limit int) (any, error)
	QueryRelations(ctx context.Context, entity string) ([]map[string]any, error)
	AddEntity(ctx context.Context, entity, entityType string, properties any) error
	AddRelation(ctx context.Context, from, to, relation string) error
	Export(ctx context.Context, path, format string) error
	Import(ctx context.Context, path string) (int, error)
	Clear(ctx context.Context) error
}
