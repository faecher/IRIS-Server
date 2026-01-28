-- IRIS Database Setup Script

-- Versioning Schema at the bottom of the file 


-- =============================================================================
-- MARK: Functions
-- =============================================================================

-- Trigger function to automatically update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS '
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
' language 'plpgsql';


-- =============================================================================
-- MARK: Trackers
-- =============================================================================

-- Organizations table (multi-tenant root)
CREATE TABLE trackers (
	-- attributes
	tracker_id uuid PRIMARY KEY,
	name text NOT NULL,
	battery SMALLINT,

	position_longitude DOUBLE PRECISION,
	position_latitude DOUBLE PRECISION,

	-- Settings & Configuration
	created_at timestamp DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_trackers_updated_at 
    BEFORE UPDATE ON trackers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();



CREATE TABLE chirpstack_trackers (
    tracker_id UUID PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
    dev_eui    TEXT NOT NULL UNIQUE
);

CREATE TABLE tetra_trackers (
    tracker_id UUID PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
    issi       TEXT NOT NULL UNIQUE
);

-- =============================================================================
-- MARK: Resources
-- =============================================================================

CREATE TABLE resources (
	resource_id uuid PRIMARY KEY,
	marker_id uuid,
	name text NOT NULL,
	type text NOT NULL,
	status SMALLINT,

	
	-- System
	created_at timestamp DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_resources_updated_at 
    BEFORE UPDATE ON resources 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- =============================================================================
-- MARK: MCP Config
-- =============================================================================

CREATE TABLE mcp_config (
	id SERIAL PRIMARY KEY,

	enabled boolean NOT NULL DEFAULT FALSE,
	url text NOT NULL,
	api_key text NOT NULL,

	operation_id UUID,
	siteplan_id UUID,

	
	-- System
	created_at timestamp DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_mcp_config_updated_at 
    BEFORE UPDATE ON mcp_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();



-- =============================================================================
-- RELATIONSHIP TABLES
-- =============================================================================

CREATE TABLE trackers_resource (
    tracker_id uuid REFERENCES trackers(tracker_id) UNIQUE ON DELETE CASCADE,
	resource_id uuid REFERENCES resources(resource_id) ON DELETE CASCADE,

	PRIMARY KEY (tracker_id, resource_id),
);

CREATE TABLE resource_marker (
	resource_id uuid REFERENCES resources(resource_id) UNIQUE ON DELETE CASCADE,
	marker_id uuid PRIMARY KEY,
	siteplan_id uuid,
);

CREATE INDEX idx_resource_marker_siteplan ON resource_marker(siteplan_id);

-- =============================================================================
-- VERSIONING SCHEMA
-- MARK: Versioning
-- =============================================================================

CREATE TABLE schema_versions (
    id SERIAL PRIMARY KEY,
    version INTEGER NOT NULL UNIQUE,
    description TEXT,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    applied_by TEXT DEFAULT CURRENT_USER
);

INSERT INTO schema_versions (version, description) VALUES 
(1, 'v1.0 - Initial schema, hopefully the last one (probably not)');