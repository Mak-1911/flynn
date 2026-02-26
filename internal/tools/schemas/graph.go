// Package schemas provides graph tool schema definitions.
package schemas

// RegisterGraphTools registers all graph tool schemas to the registry.
func RegisterGraphTools(registry *Registry) {
	registry.Register(NewSchema("graph_ingest_file", "Ingest a file into the knowledge graph").
		AddParam("path", "string", "Absolute path to the file", true).
		Build())

	registry.Register(NewSchema("graph_ingest_text", "Ingest raw text into the knowledge graph").
		AddParam("content", "string", "Text content to ingest", true).
		AddParam("path", "string", "Virtual path for the content", false).
		AddParam("title", "string", "Title for the content", false).
		Build())

	registry.Register(NewSchema("graph_entity_upsert", "Create or update an entity in the knowledge graph").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type (e.g., Person, Concept, File)", true).
		AddParam("description", "string", "Entity description", false).
		Build())

	registry.Register(NewSchema("graph_link", "Create a relationship between two entities").
		AddParam("source_name", "string", "Source entity name", true).
		AddParam("source_type", "string", "Source entity type", true).
		AddParam("target_name", "string", "Target entity name", true).
		AddParam("target_type", "string", "Target entity type", true).
		AddParam("relation_type", "string", "Type of relationship", true).
		Build())

	registry.Register(NewSchema("graph_search", "Search for entities in the knowledge graph").
		AddParam("query", "string", "Search query", true).
		Build())

	registry.Register(NewSchema("graph_relations", "Get relations for an entity").
		AddParam("entity_id", "string", "Entity ID", false).
		AddParam("name", "string", "Entity name (alternative to ID)", false).
		AddParam("type", "string", "Entity type (required with name)", false).
		Build())

	registry.Register(NewSchema("graph_related", "Get entities related to an entity").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type", true).
		Build())

	registry.Register(NewSchema("graph_summarize", "Summarize an entity and its relations").
		AddParam("name", "string", "Entity name", true).
		AddParam("type", "string", "Entity type", true).
		Build())

	registry.Register(NewSchema("graph_stats", "Get knowledge graph statistics").
		Build())
}
