// Package tool provides JSON schema definitions for OpenAI/Anthropic tool calling.
// This is draft code for review - not yet integrated into the codebase.
package tool

import (
	"encoding/json"
)

// Schema defines a tool's JSON schema for OpenAI/Anthropic formats.
type Schema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// SchemaBuilder provides a fluent interface for building tool schemas.
type SchemaBuilder struct {
	schema *Schema
}

// NewSchema creates a new schema builder with the given name and description.
func NewSchema(name, description string) *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": make(map[string]interface{}),
				"required":   make([]string, 0),
			},
		},
	}
}

// AddParam adds a parameter to the schema.
// paramType should be a JSON Schema type: "string", "number", "boolean", "array", "object"
func (b *SchemaBuilder) AddParam(name, paramType, description string, required bool) *SchemaBuilder {
	props := b.schema.Parameters["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        paramType,
		"description": description,
	}
	if required {
		req := b.schema.Parameters["required"].([]string)
		b.schema.Parameters["required"] = append(req, name)
	}
	return b
}

// AddParamWithEnum adds a parameter with an enum constraint (specific values allowed).
func (b *SchemaBuilder) AddParamWithEnum(name, paramType, description string, enum []string, required bool) *SchemaBuilder {
	props := b.schema.Parameters["properties"].(map[string]interface{})
	paramDef := map[string]interface{}{
		"type":        paramType,
		"description": description,
	}
	if len(enum) > 0 {
		paramDef["enum"] = enum
	}
	props[name] = paramDef
	if required {
		req := b.schema.Parameters["required"].([]string)
		b.schema.Parameters["required"] = append(req, name)
	}
	return b
}

// AddObjectParam adds an object parameter with nested properties.
func (b *SchemaBuilder) AddObjectParam(name, description string, properties map[string]interface{}, required []string) *SchemaBuilder {
	props := b.schema.Parameters["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        "object",
		"description": description,
		"properties":  properties,
		"required":    required,
	}
	return b
}

// Build returns the constructed schema.
func (b *SchemaBuilder) Build() *Schema {
	return b.schema
}

// Registry holds all tool schemas.
type Registry struct {
	schemas map[string]*Schema
}

// NewRegistry creates a new empty schema registry.
func NewRegistry() *Registry {
	return &Registry{schemas: make(map[string]*Schema)}
}

// Register adds a schema to the registry.
func (r *Registry) Register(schema *Schema) {
	r.schemas[schema.Name] = schema
}

// Get retrieves a schema by name.
func (r *Registry) Get(name string) (*Schema, bool) {
	s, ok := r.schemas[name]
	return s, ok
}

// List returns all registered schema names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// ToOpenAIFormat converts schemas to OpenAI function calling format.
// Returns a slice of maps suitable for inclusion in the "tools" parameter.
func (r *Registry) ToOpenAIFormat() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.schemas))
	for _, schema := range r.schemas {
		result = append(result, map[string]interface{}{
			"type":     "function",
			"function": schema,
		})
	}
	return result
}

// ToAnthropicFormat converts schemas to Anthropic tool use format.
// Returns a slice of maps suitable for inclusion in the "tools" parameter.
func (r *Registry) ToAnthropicFormat() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.schemas))
	for _, schema := range r.schemas {
		anthropicSchema := map[string]interface{}{
			"name":         schema.Name,
			"description":  schema.Description,
			"input_schema": schema.Parameters,
		}
		result = append(result, anthropicSchema)
	}
	return result
}

// ToJSON returns the registry as JSON for debugging/inspection.
func (r *Registry) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r.schemas, "", "  ")
}

// Clone creates a deep copy of the registry.
func (r *Registry) Clone() *Registry {
	clone := NewRegistry()
	for _, schema := range r.schemas {
		// Deep copy parameters
		paramsBytes, _ := json.Marshal(schema.Parameters)
		var paramsCopy map[string]interface{}
		json.Unmarshal(paramsBytes, &paramsCopy)

		clone.Register(&Schema{
			Name:        schema.Name,
			Description: schema.Description,
			Parameters:  paramsCopy,
		})
	}
	return clone
}

// Merge merges another registry into this one.
// Existing schemas with the same name are NOT overwritten.
func (r *Registry) Merge(other *Registry) {
	for name, schema := range other.schemas {
		if _, exists := r.schemas[name]; !exists {
			r.schemas[name] = schema
		}
	}
}
