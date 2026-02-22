// Package memory handles all persistent storage using SQLite.
//
// Uses two separate databases:
// - personal.db: User's private conversations and data
// - team.db: Team-shared data with tenant isolation
package memory

import (
	"database/sql"
	// SQLite driver (required for database/sql registration).
	_ "github.com/mattn/go-sqlite3"
)

// Store manages both personal and team databases.
type Store struct {
	personal *sql.DB
	team     *sql.DB
}

// Open opens both SQLite databases at the given paths.
// Creates the databases and tables if they don't exist.
func Open(personalPath, teamPath string) (*Store, error) {
	// Open personal database
	personal, err := openDB(personalPath)
	if err != nil {
		return nil, err
	}

	// Open team database
	team, err := openDB(teamPath)
	if err != nil {
		personal.Close()
		return nil, err
	}

	store := &Store{
		personal: personal,
		team:     team,
	}

	// Initialize schemas
	if err := store.initPersonal(); err != nil {
		personal.Close()
		team.Close()
		return nil, err
	}

	if err := store.initTeam(); err != nil {
		personal.Close()
		team.Close()
		return nil, err
	}

	return store, nil
}

// NewStore is an alias for Open for convenience.
func NewStore(personalPath, teamPath string) (*Store, error) {
	return Open(personalPath, teamPath)
}

// Open opens both SQLite databases at the given paths.
// Creates the databases and tables if they don't exist.
// Deprecated: Use Open or NewStore instead.
func openDBs(personalPath, teamPath string) (*Store, error) {
	// Open personal database
	personal, err := openDB(personalPath)
	if err != nil {
		return nil, err
	}

	// Open team database
	team, err := openDB(teamPath)
	if err != nil {
		personal.Close()
		return nil, err
	}

	store := &Store{
		personal: personal,
		team:     team,
	}

	// Initialize schemas
	if err := store.initPersonal(); err != nil {
		personal.Close()
		team.Close()
		return nil, err
	}

	if err := store.initTeam(); err != nil {
		personal.Close()
		team.Close()
		return nil, err
	}

	return store, nil
}

// openDB opens a single SQLite database with optimal settings.
func openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	// Set performance pragmas
	pragmas := []string{
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 30000000000",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

// Close closes both database connections.
func (s *Store) Close() error {
	var errs []error

	if s.personal != nil {
		if err := s.personal.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.team != nil {
		if err := s.team.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Personal returns the personal database connection.
func (s *Store) Personal() *sql.DB {
	return s.personal
}

// Team returns the team database connection.
func (s *Store) Team() *sql.DB {
	return s.team
}

// ============================================================
// PERSONAL DB SCHEMA
// ============================================================

func (s *Store) initPersonal() error {
	schema := `
	-- Schema version tracking
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		description TEXT
	);

	-- ============================================================
	-- CONVERSATIONS & MESSAGES (Personal)
	-- ============================================================

	CREATE TABLE IF NOT EXISTS conversations (
		id              TEXT PRIMARY KEY,
		user_id         TEXT NOT NULL,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		title           TEXT,
		summary         TEXT,
		message_count   INTEGER NOT NULL DEFAULT 0,
		is_archived     INTEGER NOT NULL DEFAULT 0,
		metadata_json   TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_conversations_user ON conversations(user_id, updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_conversations_created ON conversations(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_conversations_archived ON conversations(is_archived);

	CREATE TABLE IF NOT EXISTS messages (
		id              TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		role            TEXT NOT NULL,
		content         TEXT NOT NULL,
		tokens_used     INTEGER NOT NULL DEFAULT 0,
		cost            REAL NOT NULL DEFAULT 0,
		tier            INTEGER NOT NULL DEFAULT 0,
		model           TEXT,
		plan_id         TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		metadata_json   TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_tier ON messages(tier);

	-- ============================================================
	-- USER PROFILE
	-- ============================================================

	CREATE TABLE IF NOT EXISTS user_profile (
		id              TEXT PRIMARY KEY,
		name            TEXT,
		timezone        TEXT DEFAULT 'UTC',
		language        TEXT DEFAULT 'en',
		response_style  TEXT DEFAULT 'balanced',
		cost_sensitivity TEXT DEFAULT 'balanced',
		preferences_json TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);

	-- ============================================================
	-- COST TRACKING (Personal)
	-- ============================================================

	CREATE TABLE IF NOT EXISTS cost_history (
		id              TEXT PRIMARY KEY,
		date            TEXT NOT NULL,
		hour            INTEGER NOT NULL,
		tier            INTEGER NOT NULL,
		model           TEXT NOT NULL,
		request_type    TEXT,
		tokens_input    INTEGER NOT NULL DEFAULT 0,
		tokens_output   INTEGER NOT NULL DEFAULT 0,
		tokens_total    INTEGER NOT NULL DEFAULT 0,
		cost            REAL NOT NULL DEFAULT 0,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);

	CREATE INDEX IF NOT EXISTS idx_cost_date ON cost_history(date);
	CREATE INDEX IF NOT EXISTS idx_cost_date_hour ON cost_history(date, hour);
	CREATE INDEX IF NOT EXISTS idx_cost_tier ON cost_history(tier);

	-- ============================================================
	-- FULL-TEXT SEARCH
	-- ============================================================

	CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		content,
		content_rowid=rowid
	);

	CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON messages BEGIN
		INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON messages BEGIN
		DELETE FROM messages_fts WHERE rowid = OLD.rowid;
	END;

	CREATE TRIGGER IF NOT EXISTS messages_fts_update AFTER UPDATE ON messages BEGIN
		UPDATE messages_fts SET content = NEW.content WHERE rowid = NEW.rowid;
	END;

	-- ============================================================
	-- TRIGGERS
	-- ============================================================

	CREATE TRIGGER IF NOT EXISTS conversations_updated
		AFTER UPDATE ON conversations
		BEGIN
			UPDATE conversations SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS messages_count_insert
		AFTER INSERT ON messages
		BEGIN
			UPDATE conversations
			SET message_count = message_count + 1, updated_at = strftime('%s', 'now')
			WHERE id = NEW.conversation_id;
		END;

	CREATE TRIGGER IF NOT EXISTS messages_count_delete
		AFTER DELETE ON messages
		BEGIN
			UPDATE conversations
			SET message_count = message_count - 1, updated_at = strftime('%s', 'now')
			WHERE id = OLD.conversation_id;
		END;
	`

	_, err := s.personal.Exec(schema)
	if err != nil {
		return err
	}

	if err := ensureSchemaVersion(s.personal, 1, "Initial personal schema"); err != nil {
		return err
	}

	return nil
}

// ============================================================
// TEAM DB SCHEMA
// ============================================================

func (s *Store) initTeam() error {
	schema := `
	-- Schema version tracking
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		description TEXT
	);

	-- ============================================================
	-- TENANTS
	-- ============================================================

	CREATE TABLE IF NOT EXISTS tenants (
		id              TEXT PRIMARY KEY,
		name            TEXT NOT NULL,
		settings_json   TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);

	-- ============================================================
	-- TEAM MEMBERS
	-- ============================================================

	CREATE TABLE IF NOT EXISTS team_members (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		name            TEXT NOT NULL,
		role            TEXT NOT NULL,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_members_tenant ON team_members(tenant_id);

	-- ============================================================
	-- TEAM CONVERSATIONS & MESSAGES
	-- ============================================================

	CREATE TABLE IF NOT EXISTS team_conversations (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		title           TEXT,
		summary         TEXT,
		message_count   INTEGER NOT NULL DEFAULT 0,
		is_archived     INTEGER NOT NULL DEFAULT 0,
		metadata_json   TEXT,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_team_conv_tenant ON team_conversations(tenant_id, updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_team_conv_created ON team_conversations(created_at DESC);

	CREATE TABLE IF NOT EXISTS team_messages (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		conversation_id TEXT NOT NULL,
		user_id         TEXT NOT NULL,
		role            TEXT NOT NULL,
		content         TEXT NOT NULL,
		tokens_used     INTEGER NOT NULL DEFAULT 0,
		cost            REAL NOT NULL DEFAULT 0,
		tier            INTEGER NOT NULL DEFAULT 0,
		model           TEXT,
		plan_id         TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		metadata_json   TEXT,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (conversation_id) REFERENCES team_conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_team_msg_conv ON team_messages(conversation_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_team_msg_tenant ON team_messages(tenant_id, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_team_msg_user ON team_messages(user_id, created_at DESC);

	-- ============================================================
	-- SHARED KNOWLEDGE GRAPH
	-- ============================================================

	CREATE TABLE IF NOT EXISTS team_entities (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		name            TEXT NOT NULL,
		entity_type     TEXT NOT NULL,
		description     TEXT,
		metadata_json   TEXT,
		embedding_id    TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		importance      REAL DEFAULT 0,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_team_entities_tenant ON team_entities(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_team_entities_name ON team_entities(tenant_id, name);
	CREATE INDEX IF NOT EXISTS idx_team_entities_type ON team_entities(tenant_id, entity_type);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_team_entities_unique ON team_entities(tenant_id, name, entity_type);

	CREATE TABLE IF NOT EXISTS team_relations (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		source_id       TEXT NOT NULL,
		target_id       TEXT NOT NULL,
		relation_type   TEXT NOT NULL,
		metadata_json   TEXT,
		confidence      REAL DEFAULT 1,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, source_id, target_id, relation_type)
	);

	CREATE INDEX IF NOT EXISTS idx_team_relations_tenant ON team_relations(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_team_relations_source ON team_relations(tenant_id, source_id);
	CREATE INDEX IF NOT EXISTS idx_team_relations_target ON team_relations(tenant_id, target_id);

	-- ============================================================
	-- SHARED PLAN LIBRARY
	-- ============================================================

	CREATE TABLE IF NOT EXISTS team_plans (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		intent_category TEXT NOT NULL,
		description     TEXT NOT NULL,
		steps_json      TEXT NOT NULL,
		variables_json  TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		is_active       INTEGER NOT NULL DEFAULT 1,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_team_plans_tenant_intent ON team_plans(tenant_id, intent_category);
	CREATE INDEX IF NOT EXISTS idx_team_plans_active ON team_plans(is_active);

	CREATE TABLE IF NOT EXISTS team_plan_patterns (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		intent_category TEXT NOT NULL,
		plan_id         TEXT NOT NULL,
		usage_count     INTEGER NOT NULL DEFAULT 0,
		success_count   INTEGER NOT NULL DEFAULT 0,
		failure_count   INTEGER NOT NULL DEFAULT 0,
		success_rate    REAL NOT NULL DEFAULT 0,
		last_used       INTEGER,
		last_succeeded  INTEGER,
		last_failed     INTEGER,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, intent_category)
	);

	CREATE INDEX IF NOT EXISTS idx_team_patterns_tenant ON team_plan_patterns(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_team_patterns_success ON team_plan_patterns(tenant_id, success_rate DESC);
	CREATE INDEX IF NOT EXISTS idx_team_patterns_usage ON team_plan_patterns(tenant_id, usage_count DESC);

	CREATE TABLE IF NOT EXISTS team_plan_executions (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		plan_id         TEXT NOT NULL,
		pattern_id      TEXT,
		variables_json  TEXT,
		status          TEXT NOT NULL,
		error_message   TEXT,
		started_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		completed_at    INTEGER,
		duration_ms     INTEGER,
		total_tokens    INTEGER DEFAULT 0,
		total_cost      REAL DEFAULT 0,
		step_count      INTEGER NOT NULL,
		steps_completed INTEGER NOT NULL DEFAULT 0,
		steps_json      TEXT,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_team_executions_tenant ON team_plan_executions(tenant_id, started_at DESC);

	-- ============================================================
	-- SHARED DOCUMENTS
	-- ============================================================

	CREATE TABLE IF NOT EXISTS team_documents (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		path            TEXT NOT NULL,
		title           TEXT,
		content_preview TEXT,
		file_type       TEXT,
		size_bytes      INTEGER,
		language        TEXT,
		indexed_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		chunk_count     INTEGER DEFAULT 0,
		metadata_json   TEXT,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, path)
	);

	CREATE INDEX IF NOT EXISTS idx_team_docs_tenant ON team_documents(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_team_docs_type ON team_documents(tenant_id, file_type);

	CREATE TABLE IF NOT EXISTS team_doc_chunks (
		id              TEXT PRIMARY KEY,
		tenant_id       TEXT NOT NULL,
		document_id     TEXT NOT NULL,
		chunk_index     INTEGER NOT NULL,
		content         TEXT NOT NULL,
		metadata_json   TEXT,
		created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, document_id, chunk_index)
	);

	CREATE INDEX IF NOT EXISTS idx_team_chunks_doc ON team_doc_chunks(document_id, chunk_index);

	-- ============================================================
	-- TRIGGERS
	-- ============================================================

	CREATE TRIGGER IF NOT EXISTS team_conversations_updated
		AFTER UPDATE ON team_conversations
		BEGIN
			UPDATE team_conversations SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS team_entities_updated
		AFTER UPDATE ON team_entities
		BEGIN
			UPDATE team_entities SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS team_relations_updated
		AFTER UPDATE ON team_relations
		BEGIN
			UPDATE team_relations SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
		END;

	CREATE TRIGGER IF NOT EXISTS team_messages_count_insert
		AFTER INSERT ON team_messages
		BEGIN
			UPDATE team_conversations
			SET message_count = message_count + 1, updated_at = strftime('%s', 'now')
			WHERE id = NEW.conversation_id;
		END;

	CREATE TRIGGER IF NOT EXISTS team_plan_execution_complete
		AFTER UPDATE ON team_plan_executions WHEN NEW.status = 'completed'
		BEGIN
			UPDATE team_plan_patterns
			SET usage_count = usage_count + 1,
				success_count = success_count + 1,
				success_rate = CAST(success_count AS REAL) / usage_count,
				last_used = strftime('%s', 'now'),
				last_succeeded = strftime('%s', 'now'),
				updated_at = strftime('%s', 'now')
			WHERE id = NEW.pattern_id;
		END;

	CREATE TRIGGER IF NOT EXISTS team_plan_execution_failed
		AFTER UPDATE ON team_plan_executions WHEN NEW.status = 'failed'
		BEGIN
			UPDATE team_plan_patterns
			SET usage_count = usage_count + 1,
				failure_count = failure_count + 1,
				success_rate = CAST(success_count AS REAL) / usage_count,
				last_used = strftime('%s', 'now'),
				last_failed = strftime('%s', 'now'),
				updated_at = strftime('%s', 'now')
			WHERE id = NEW.pattern_id;
		END;
	`

	_, err := s.team.Exec(schema)
	if err != nil {
		return err
	}

	if err := ensureSchemaVersion(s.team, 1, "Initial team schema"); err != nil {
		return err
	}

	return nil
}

func ensureSchemaVersion(db *sql.DB, version int, description string) error {
	var current sql.NullInt64
	if err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&current); err != nil {
		return err
	}

	if !current.Valid || int(current.Int64) < version {
		_, err := db.Exec(
			"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
			version,
			description,
		)
		return err
	}

	return nil
}
