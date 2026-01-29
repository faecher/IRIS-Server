# IRIS-Server Go Rewrite Analysis

**Date:** January 28, 2026  
**Status:** In Development

## Executive Summary

This document summarizes the rewrite of IRIS-Server from Python/FastAPI to Go/Gin, including all outstanding TODOs, functional differences, API route changes, new features, and features left out.

---

## 1. Outstanding TODOs

### 1.1 Tracker Management
- **[internal/repository/trackerCreation.go](../internal/repository/trackerCreation.go)
  - Implement `CreateTETRATracker()` function body as soon as we know how to
  - Finish the entire TETRA implementation

### 1.2 MCP Integration
- **[internal/mcp_control/positionMirror.go](../internal/mcp_control/positionMirror.go)**
  - Determine appropriate icon for marker (currently hardcoded as "BASIC_PIN") - should we read from resource type?

- **[cmd/iris-server/iris-server.go](../cmd/iris-server/iris-server.go)**
  - Implement periodic check of MCP resources to update status and resource info in database

### 1.3 System Information
- **[internal/handlers/system.go](../internal/handlers/system.go)**
  - Read version from build-time variable or config
  - Return version information in response

---

## 2. API Routes Comparison

### 2.1 Routes Unchanged
Routes that exist in both Python and Go with the same path and method:

- `GET /tracker/` - List all trackers
- `GET /resources/` - List all MCP resources  
- `GET /system/status` - System health status
- `GET /system/version` - Application version
- `GET /mcp/operations` - Fetch MCP operations from API
- `POST /mcp/start` - Configure and enable MCP integration

### 2.2 Routes Changed

#### Gateway/Webhook Routes
- **Python:** `POST /gateway/data`
- **Go:** `POST /chirpstackGateway/data`
- **Reason:** More explicit naming to differentiate future tracker types (TETRA, etc.)

#### Tracker Assignment Routes
- **Python:** `POST /tracker/{instance_id}`
  - Body: `TrackerUpdateModel` with resource field
  - Uses integer IDs
- **Go:** `POST /tracker/assign/{tracker_id}/{resource_id}`
  - Resource ID now in URL path
  - Support for unassignment via empty resource_id
  - Uses UUID instead of integer IDs
- **Reason:** More RESTful design, clearer intent, supports resource unassignment

#### MCP Configuration
- **Python:** `GET /mcp/config` returns operation info with `operation_selected` and `operation` fields
- **Go:** `GET /mcp/config` returns only basic config (enabled, api_key, url, selected operation and selected siteplan). The differenciation between operation and operation_selected got removed.
- **Reason:** Simpler design that only stores data that is necessary for IRIS functionality and queries needed data from MCP dynamically.

### 2.3 New Routes in Go
| Route | Method | Purpose |
|-------|--------|---------|
| `/tracker/rename/{tracker_id}` | POST | Rename tracker with new name in body |
| `/mcp/operations/set/{id}` | POST | Select MCP operation (replaces enable/disable) |
| `/mcp/siteplans` | GET | Retrieve MCP siteplans for selected operation |
| `/mcp/siteplans/set/{id}` | POST | Select active siteplan for marker placement |

### 2.4 Deleted Routes
Routes that existed in Python but were removed in Go:

- `POST /mcp/operations/enable` 
- `POST /mcp/operations/disable`
Operation enable/disable should be handled in MCP itself. We now only select an operation to be able to display siteplan options.

---

## 3. Database Schema Differences

### 3.1 Simplifications & Improvements

#### UUID Consolidation
- **Python:** Resources have dual identifiers (`id` as integer, `uid` as string UUID)
- **Go:** Single `resource_id` UUID primary key
- **Benefit:** Cleaner schema, eliminates confusion between internal and external IDs

#### Tracker Type Separation
- **Python:** Single `trackers` table with `deviceEUI` field
- **Go:** 
  - Base `trackers` table with common fields
  - Separate `chirpstack_trackers` table with `dev_eui`
  - Separate `tetra_trackers` table with `issi`
- **Benefit:** Better type safety, easier extensibility for new tracker types

#### Position Storage
- **Python:** `long` and `lat` float columns
- **Go:** `position_longitude` and `position_latitude` double precision
- **Benefit:** More descriptive names, higher precision

#### Timestamps
- **Python:** `lastUpdated` as unix timestamp integer
- **Go:** `created_at` and `updated_at` as PostgreSQL timestamp with automatic triggers
- **Benefit:** Native PostgreSQL types, automatic update tracking

#### Tracker-Resource Relationship
- **Python:** Direct foreign key `resourceId` in tracker table
- **Go:** Separate junction table `trackers_resource` with `tracker_id` and `resource_id`
  - One-to-many relationship enforced via UNIQUE constraint on `tracker_id`
  - Supports ON CONFLICT for upsert operations
- **Benefit:** Go version uses cleaner normalized schema, easier to extend to many-to-many if needed

### 3.2 New Tables in Go
- `mcp_config` - Stores MCP server configuration (enabled, url, api_key, operation_id, siteplan_id)
- `trackers_resource` - Junction table for tracker-resource assignments (replaces Python's direct foreign key)
- `resource_marker` - Stores marker IDs per resource per siteplan (enables multiple markers for different siteplans)

### 3.3 Removed Fields
- **Python `operations` table:** Not yet implemented in Go (TODO: determine if needed)
  - Fields: `id`, `uid`, `title`, `active`, `archived`, `selected`
  - **Status:** Functionality questioned in TODO comment

---

## 4. Functional Differences

### 4.1 Architecture & Code Structure

#### Tracker Abstraction
- **Python:** Concrete `Tracker` class with SQLAlchemy ORM
- **Go:** 
  - `Tracker` interface with `GetPosition()`, `GetLastUpdate()`, `GetBattery()`
  - `BaseTracker` struct with common fields
  - Type-specific trackers extend BaseTracker
- **Benefit:** True polymorphism, easier to add new tracker types

#### Configuration Management
- **Python:** `Settings` class loaded via Pydantic settings
- **Go:** `Config` struct loaded via `caarlos0/env` with structured sub-configs from environment variables
- **Benefit:** More granular organization (MCPConfig, SQLConfig, SentryConfig), config changable from docker-compose file

### 4.2 Chirpstack Message Handling

#### Update Logic
- **Python:** Compares message timestamp against `lastUpdated` before applying updates
- **Go:** Currently missing timestamp comparison (TODO in implementation)
- **Impact:** Go version may need to add timestamp validation - is that actually a problem?

### 4.3 MCP Integration

#### Position Mirroring
- **Python:** 
  - Background task via `get_mcp_data()` scheduled function
  - Updates resources from MCP tableau
- **Go:**
  - Direct HTTP client in `internal/mcp_control/positionMirror.go`
  - `UpdateTrackerInMCP()` for pushing tracker positions to markers
  - **Multi-siteplan support:** Stores one marker ID per resource per siteplan in `resource_marker` table
  - Automatically uses currently selected siteplan from `mcp_config`
  - Creates new marker if none exists for resource+siteplan combination
  - Updates existing marker if found
  - Sync from MCP to Go still missing (TODO: periodic resource updates)
- **Benefit:** Go version supports multiple siteplans with separate markers, Python limited to single marker per resource

#### Operations and Siteplans Management
- **Python:** 
  - Fetches operations from MCP on startup
  - Stores in database with `selected` flag
  - Filters resources by selected operation
- **Go:** 
  - Fetches operations dynamically via `/mcp/operations` endpoint
  - Stores selected operation ID in `mcp_config.operation_id`
  - Fetches siteplans based on operation's place ID
  - Stores selected siteplan ID in `mcp_config.siteplan_id`
  - Validates operations and siteplans exist in MCP before saving
  - **New workflow:** Select operation → Get associated place → Fetch siteplans for place → Select siteplan
- **Benefit:** Go version has cleaner API design with separate operation/siteplan selection, validation against live MCP data


### 4.4 Concurrency

#### Request Handling
- **Python:** FastAPI async/await with event loop
- **Go:** Goroutines per request (automatic via `http.Server`)
- **Benefit:** Go's native concurrency is simpler, no async/await needed

#### Background Tasks
- **Python:** Scheduled via external task scheduler, `async` functions
- **Go:** Not yet implemented (TODO: periodic MCP resource checks)
- **Impact:** Go version needs background goroutine with ticker

---

## 5. New Features in Go

### 5.1 Enhanced Type Safety
- UUID types throughout (`gofrs/uuid/v5`)
- Nullable UUIDs via pointer types (`*uuid.UUID`)
- Interface-based tracker abstraction

### 5.2 Improved Extensibility
- Clear separation of tracker types (Chirpstack, TETRA)
- Repository pattern for database operations
- Centralized message processing in `internal/chirpstack/messageProcessing.go`

### 5.3 Tracker Renaming API
- New endpoint: `POST /tracker/rename/{tracker_id}`
- Allows changing tracker display name
- Not present in Python version

### 5.4 MCP Configuration Endpoint
- New endpoint: `GET /mcp/config`
- Returns current MCP server configuration
- Supports retrieving enabled status, URL, and API key

### 5.5 Sentry Integration
- Built-in error tracking with Sentry SDK
- Configured via `SentryConfig` with DSN and environment

### 5.6 Resource Unassignment
- `POST /tracker/assign/{tracker_id}/` (empty resource_id)
- Explicit API for removing resource assignments
- Python version requires `None` in request body

### 5.7 Database Triggers
- Automatic `updated_at` timestamp updates via PostgreSQL triggers
- Python relies on application-level timestamp management

### 5.8 Multi-Siteplan Marker Support
- **New mechanic:** One marker ID stored per resource per siteplan
- `resource_marker` table with composite key (resource_id, siteplan_id)
- When updating tracker position, system:
  1. Retrieves currently selected siteplan from `mcp_config`
  2. Looks up marker ID for resource+siteplan combination
  3. Creates new marker if none exists, updates existing marker otherwise
- **Benefit:** Resources can have different markers on different siteplans
- **Use case:** Same vehicle tracked on multiple operation siteplans simultaneously

---

## 6. Features Left Out

### 6.1 MCP Operations Database Table
- **Status:** Not implemented - operations queried dynamically from MCP API
- **Design Decision:** Store only selected operation ID in `mcp_config`, not full operations table
- **Python Feature:** Stores operations locally in database with `active`/`selected` flags
- **Benefit:** Go version always has current data from MCP, no sync issues with stale local operations

### 6.2 Background Resource Sync
- **Status:** TODO in main.go
- **Python Feature:** Scheduled `get_mcp_data()` task pulls resources from MCP
- **Impact:** Go version needs goroutine with periodic execution

### 6.4 FastAPI Automatic Documentation
- **Status:** No equivalent in Go/Gin
- **Python Feature:** Auto-generated OpenAPI/Swagger docs
- **Impact:** May need manual API documentation or Swagger annotations


---

## 7. Migration Considerations

### 7.1 API Compatibility
- **Frontend Changes Needed:**
  - Update Chirpstack webhook URL: `/gateway/data` → `/chirpstackGateway/data`
  - Update tracker assignment endpoint: `/tracker/{id}` → `/tracker/assign/{tracker_id}/{resource_id}`
  - Handle UUID strings instead of integers for IDs
  - New tracker rename endpoint available

### 7.2 Configuration Changes
- **Environment Variables:**
  - MCP configuration now includes `MCP_ENABLED`, `MCP_SERVER_URL`, `MCP_API_KEY`
  - SQL configuration more granular: `SQL_HOST`, `SQL_PORT`, `SQL_USER`, `SQL_PASSWORD`, `SQL_DB_NAME`
  - New: `SENTRY_DSN`, `SENTRY_ENVIRONMENT`

---

## 9. Next Steps

### 9.1 Critical Path
1. ✅ Complete Chirpstack message parsing (battery, lat/long) - **DONE**
2. ✅ Implement Chirpstack webhook handler structure - **DONE** (but broken due to bugs above)
3. ✅ Implement tracker rename functionality - **DONE**
4. ✅ Implement MCP position sync (POST/PUT markers) - **DONE**
5. ✅ Implement MCP operations selection - **DONE**
6. ✅ Implement MCP siteplans fetching and selection - **DONE**
7. ✅ Implement marker-per-siteplan storage mechanism - **DONE**
8.  ✅ Save marker ID to resource in database - **DONE** (UpdateMarkerIDForResource)
9.  ✅ Resolve siteplanId requirement - **DONE** (stored in mcp_config, used from resource_marker)
10. 🔲 Add background goroutine for periodic MCP resource sync
11. 🔲 Complete system version endpoint
12. 🔲 Testing with real Chirpstack webhooks and MCP integration (cannot test until tracker creation bugs fixed)

### 9.2 Future Enhancements
- TETRA tracker implementation
- API documentation (Swagger/OpenAPI)
- Database migration scripts
- Unit and integration tests
- Deployment configuration (Docker, docker-compose)

---

## 10. Open Questions

1. **Marker Icons:** Should icon type be determined from resource type, or remain hardcoded as "BASIC_PIN"?
2. **Version Info:** Should version be build-time variable, config file, or git tag?

---

## Document Status
- **Created:** January 28, 2026
- **Last Updated:** January 28, 2026 (Verified and updated with critical bug discoveries)
- **Maintained By:** Development team
- **Review Schedule:** Update after major milestones
