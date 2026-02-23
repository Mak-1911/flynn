// Package graph provides context retrieval for prompts.
package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flynn-ai/flynn/internal/memory"
)

// ContextBuilder builds compact knowledge graph context strings.
type ContextBuilder struct {
	Store        *memory.GraphStore
	MaxEntities  int
	MaxRelations int
	MaxChars     int
}

// FromText builds a context string by searching entities related to the text.
func (c *ContextBuilder) FromText(ctx context.Context, tenantID, text string) (string, error) {
	if c == nil || c.Store == nil {
		return "", fmt.Errorf("context builder not initialized")
	}

	query := keywords(text)
	if query == "" {
		return "", nil
	}

	entities, err := c.Store.SearchEntities(ctx, tenantID, query, c.maxEntities())
	if err != nil {
		return "", err
	}
	if len(entities) == 0 {
		return "", nil
	}

	relations := make([]*memory.Relation, 0)
	relSeen := make(map[string]bool)
	for _, e := range entities {
		rels, err := c.Store.GetRelations(ctx, tenantID, e.ID, c.maxRelations())
		if err != nil {
			continue
		}
		for _, r := range rels {
			key := r.ID
			if relSeen[key] {
				continue
			}
			relSeen[key] = true
			relations = append(relations, r)
		}
	}

	nameCache := map[string]string{}
	for _, e := range entities {
		nameCache[e.ID] = e.Name
	}
	textOut := formatContext(ctx, c.Store, tenantID, entities, relations, nameCache, c.maxChars())
	return textOut, nil
}

func (c *ContextBuilder) maxEntities() int {
	if c.MaxEntities <= 0 {
		return 10
	}
	return c.MaxEntities
}

func (c *ContextBuilder) maxRelations() int {
	if c.MaxRelations <= 0 {
		return 10
	}
	return c.MaxRelations
}

func (c *ContextBuilder) maxChars() int {
	if c.MaxChars <= 0 {
		return 1200
	}
	return c.MaxChars
}

func formatContext(ctx context.Context, store *memory.GraphStore, tenantID string, entities []*memory.Entity, relations []*memory.Relation, nameCache map[string]string, maxChars int) string {
	var b strings.Builder
	b.WriteString("Entities:\n")
	for _, e := range entities {
		line := fmt.Sprintf("- %s (%s)\n", e.Name, e.EntityType)
		if b.Len()+len(line) > maxChars {
			break
		}
		b.WriteString(line)
	}

	if len(relations) > 0 {
		b.WriteString("Relations:\n")
		for _, r := range relations {
			src := resolveName(ctx, store, tenantID, r.SourceID, nameCache)
			dst := resolveName(ctx, store, tenantID, r.TargetID, nameCache)
			line := fmt.Sprintf("- %s --%s--> %s\n", src, r.RelationType, dst)
			if b.Len()+len(line) > maxChars {
				break
			}
			b.WriteString(line)
		}
	}

	return b.String()
}

func resolveName(ctx context.Context, store *memory.GraphStore, tenantID, id string, nameCache map[string]string) string {
	if name, ok := nameCache[id]; ok {
		return name
	}
	if store == nil {
		return id
	}
	entity, err := store.GetEntityByID(ctx, tenantID, id)
	if err == nil && entity != nil {
		nameCache[id] = entity.Name
		return entity.Name
	}
	return id
}

func keywords(text string) string {
	words := strings.Fields(strings.ToLower(text))
	seen := make(map[string]bool)
	var tokens []string
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(w) < 4 {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		tokens = append(tokens, w)
	}
	sort.Strings(tokens)
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, " ")
}
