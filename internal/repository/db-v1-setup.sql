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

-- Trigger function to clear tracker assignments when operation changes
CREATE OR REPLACE FUNCTION clear_tracker_assignments_on_operation_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.operation_id IS DISTINCT FROM NEW.operation_id THEN
        DELETE FROM trackers_resource;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


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

-- Note: updated_at is manually controlled to prevent buffered/old messages from updating the timestamp



CREATE TABLE chirpstack_trackers (
    tracker_id UUID PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
    dev_eui    TEXT NOT NULL UNIQUE
);

CREATE TABLE tetra_trackers (
    tracker_id UUID PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
    issi       TEXT NOT NULL UNIQUE
);

CREATE TABLE traccar_trackers (
	tracker_id UUID PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
	traccar_id BIGINT NOT NULL UNIQUE
);

-- =============================================================================
-- MARK: Resources
-- =============================================================================

CREATE TABLE resources (
	resource_id uuid PRIMARY KEY,
	name text NOT NULL,
	type text NOT NULL,
	
	-- System
	created_at timestamp DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_resources_updated_at 
    BEFORE UPDATE ON resources 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE tableau_resources (
	tableau_resource_id uuid PRIMARY KEY,
	resource_id uuid NOT NULL REFERENCES resources(resource_id) ON DELETE CASCADE,
	operation_id uuid NOT NULL,
	status SMALLINT NOT NULL,

	-- System
	created_at timestamp DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP,

	UNIQUE (resource_id, operation_id)
);

CREATE TRIGGER update_tableau_resources_updated_at 
    BEFORE UPDATE ON tableau_resources 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- =============================================================================
-- MARK: MCP Config
-- =============================================================================

CREATE TABLE mcp_config (
	id SERIAL PRIMARY KEY,

	delete_markers_on_unassign boolean NOT NULL DEFAULT FALSE,
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

CREATE TRIGGER clear_tracker_assignments 
    BEFORE UPDATE ON mcp_config 
    FOR EACH ROW EXECUTE FUNCTION clear_tracker_assignments_on_operation_change();



-- =============================================================================
-- RELATIONSHIP TABLES
-- =============================================================================

CREATE TABLE trackers_resource (
    tracker_id uuid PRIMARY KEY REFERENCES trackers(tracker_id) ON DELETE CASCADE,
	tableau_resource_id uuid REFERENCES tableau_resources(tableau_resource_id) ON DELETE CASCADE
);

CREATE TABLE resource_marker (
	resource_id uuid REFERENCES resources(resource_id) ON DELETE CASCADE,
	marker_id uuid PRIMARY KEY,
	siteplan_id uuid,

	UNIQUE (resource_id, siteplan_id)
);

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