# Shared Tool Schema Infrastructure

Common types and utilities for defining tool schemas used across all Flynn agents.

## Overview

This package provides:
- **Schema**: Core type for defining tool JSON schemas
- **SchemaBuilder**: Fluent API for constructing tool definitions
- **Registry**: Container for managing tool collections

## Usage

```go
import "github.com/flynn-ai/flynn/drafts/tools/shared"

// Create a new tool schema
schema := shared.NewSchema("my_tool", "Does something useful").
    AddParam("path", "string", "Path to process", true).
    AddParam("recursive", "boolean", "Process recursively", false).
    Build()

// Register in a registry
registry := shared.NewRegistry()
registry.Register(schema)

// Export to different formats
openAIFormat := registry.ToOpenAIFormat()
anthropicFormat := registry.ToAnthropicFormat()
```

## Schema Format

Tool schemas follow the OpenAI function calling format:
- `name`: Tool identifier (e.g., "file_read")
- `description`: Human-readable purpose
- `parameters`: JSON Schema for input validation

## Supported Parameter Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text value | `"hello"` |
| `number` | Numeric value | `42`, `3.14` |
| `boolean` | True/false | `true` |
| `array` | List of values | `["a", "b"]` |
| `object` | Nested properties | `{"key": "value"}` |

## Naming Conventions

Tools follow the pattern `{agent}_{action}`:
- `file_read` - File agent, read action
- `code_analyze` - Code agent, analyze action
- `system_open_app` - System agent, open app action

## Integration

Each agent package imports `shared` for schema definitions:
- `file` - File system operations
- `code` - Code analysis and git operations
- `system` - System-level operations
- `task` - Task management
- `graph` - Knowledge graph operations
- `research` - Web search and fetching
