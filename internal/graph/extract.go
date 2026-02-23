// Package graph provides knowledge graph extraction utilities.
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/flynn-ai/flynn/internal/model"
)

// ExtractedEntity is a normalized entity from text.
type ExtractedEntity struct {
	Name        string
	Type        string
	Description string
}

// ExtractedRelation is a normalized relation from text.
type ExtractedRelation struct {
	SourceName string
	SourceType string
	TargetName string
	TargetType string
	Relation   string
}

// Extractor extracts entities and relations from text.
type Extractor interface {
	Extract(ctx context.Context, text string) ([]ExtractedEntity, []ExtractedRelation, error)
}

// FallbackExtractor tries primary, then falls back if it errors.
type FallbackExtractor struct {
	Primary  Extractor
	Fallback Extractor
}

// Extract implements Extractor with fallback behavior.
func (f *FallbackExtractor) Extract(ctx context.Context, text string) ([]ExtractedEntity, []ExtractedRelation, error) {
	if f.Primary != nil {
		entities, relations, err := f.Primary.Extract(ctx, text)
		if err == nil {
			return entities, relations, nil
		}
	}
	if f.Fallback != nil {
		return f.Fallback.Extract(ctx, text)
	}
	return nil, nil, nil
}

// RuleBasedExtractor extracts basic entities/relations using regex patterns.
type RuleBasedExtractor struct {
	MaxEntities  int
	MaxRelations int
}

// Extract extracts entities and relations from text using regex patterns.
func (r *RuleBasedExtractor) Extract(ctx context.Context, text string) ([]ExtractedEntity, []ExtractedRelation, error) {
	entities := r.extractEntities(text)
	relations := r.extractRelations(text)
	return entities, relations, nil
}

func (r *RuleBasedExtractor) extractEntities(text string) []ExtractedEntity {
	maxEntities := r.MaxEntities
	if maxEntities <= 0 {
		maxEntities = 20
	}

	seen := make(map[string]bool)
	var entities []ExtractedEntity

	// Proper-noun phrases.
	proper := regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`)
	for _, match := range proper.FindAllString(text, -1) {
		name := strings.TrimSpace(match)
		if len(name) < 3 {
			continue
		}
		key := "proper:" + strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		entities = append(entities, ExtractedEntity{Name: name, Type: "proper_noun"})
		if len(entities) >= maxEntities {
			return entities
		}
	}

	// Email addresses.
	email := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	for _, match := range email.FindAllString(text, -1) {
		key := "email:" + strings.ToLower(match)
		if seen[key] {
			continue
		}
		seen[key] = true
		entities = append(entities, ExtractedEntity{Name: match, Type: "email"})
		if len(entities) >= maxEntities {
			return entities
		}
	}

	// URLs.
	url := regexp.MustCompile(`https?://\S+`)
	for _, match := range url.FindAllString(text, -1) {
		key := "url:" + strings.ToLower(match)
		if seen[key] {
			continue
		}
		seen[key] = true
		entities = append(entities, ExtractedEntity{Name: match, Type: "url"})
		if len(entities) >= maxEntities {
			return entities
		}
	}

	return entities
}

func (r *RuleBasedExtractor) extractRelations(text string) []ExtractedRelation {
	maxRelations := r.MaxRelations
	if maxRelations <= 0 {
		maxRelations = 20
	}

	var relations []ExtractedRelation
	type pattern struct {
		re  *regexp.Regexp
		rel string
	}

	patterns := []pattern{
		{regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\s+is\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`), "is_a"},
		{regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\s+uses\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`), "uses"},
		{regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\s+depends on\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`), "depends_on"},
		{regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\s+works with\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`), "works_with"},
		{regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\s+member of\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+){0,2})\b`), "member_of"},
	}

	for _, p := range patterns {
		matches := p.re.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			relations = append(relations, ExtractedRelation{
				SourceName: m[1],
				SourceType: "proper_noun",
				TargetName: m[2],
				TargetType: "proper_noun",
				Relation:   p.rel,
			})
			if len(relations) >= maxRelations {
				return relations
			}
		}
	}

	return relations
}

// LLMExtractor uses a model to extract entities and relations.
type LLMExtractor struct {
	Model model.Model
}

// Extract extracts entities and relations using an LLM response.
func (l *LLMExtractor) Extract(ctx context.Context, text string) ([]ExtractedEntity, []ExtractedRelation, error) {
	if l == nil || l.Model == nil || !l.Model.IsAvailable() {
		return nil, nil, fmt.Errorf("llm extractor not available")
	}

	prompt := fmt.Sprintf(`Extract entities and relations from the text. Return ONLY JSON:
{
  "entities": [{"name": "Name", "type": "type", "description": "optional"}],
  "relations": [{"source": "Name", "source_type": "type", "target": "Name", "target_type": "type", "relation": "type"}]
}

Text:
%s
`, text)

	resp, err := l.Model.Generate(ctx, &model.Request{Prompt: prompt, JSON: true})
	if err != nil {
		return nil, nil, err
	}

	var parsed struct {
		Entities []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"entities"`
		Relations []struct {
			Source     string `json:"source"`
			SourceType string `json:"source_type"`
			Target     string `json:"target"`
			TargetType string `json:"target_type"`
			Relation   string `json:"relation"`
		} `json:"relations"`
	}

	if err := decodeJSON(resp.Text, &parsed); err != nil {
		return nil, nil, err
	}

	var entities []ExtractedEntity
	for _, e := range parsed.Entities {
		entities = append(entities, ExtractedEntity{
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
		})
	}

	var relations []ExtractedRelation
	for _, r := range parsed.Relations {
		relations = append(relations, ExtractedRelation{
			SourceName: r.Source,
			SourceType: r.SourceType,
			TargetName: r.Target,
			TargetType: r.TargetType,
			Relation:   r.Relation,
		})
	}

	return entities, relations, nil
}

func decodeJSON(text string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(text))
	return decoder.Decode(target)
}
