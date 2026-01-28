# IRIS-Server Go Rewrite Analysis

**Date:** January 28, 2026  
**Status:** In Development

## Executive Summary

This document summarizes the rewrite of IRIS-Server from Python/FastAPI to Go/Gin, including all outstanding TODOs, functional differences, API route changes, new features, and features left out.

---

## 1. Outstanding TODOs

### 1.1 Tracker Management
- **[internal/repository/trackerCreation.go](../internal/repository/trackerCreation.go)
  - Implement `CreateChirpstackTracker()` function body
  - Implement `CreateTETRATracker()` function body
  - Note: This file requires editing for each new tracker type addition

- **[internal/repository/trackers.go](../internal/repository/trackers.go)**
  - Adjust `GetAllTrackers()` to support additional tracker types beyond Chirpstack

### 1.2 MCP Integration
- **[internal/handlers/mcp.go](../internal/handlers/mcp.go)**
  - **DECISION NEEDED:** Do we need operations enable/disable functionality? We just want to move pins in MCP
  - Query operations table filtered by `active=true` for `getMCPOperations()`
  - Parse `MCPOperationConfig` from request body (contains uid) for enable/disable operations
  - Find operation by uid in database
  - Set `operation.selected = true/false`
  - Commit changes to database
  - Return `{"status": 200}` responses
  - Query for active AND selected operation in `getMCPConfig()`
  - Build response object including `operation_selected` and `operation` (uid)

- **[internal/mcp_control/positionMirror.go](../internal/mcp_control/positionMirror.go)**
  - Determine appropriate icon for marker (currently hardcoded as "BASIC_PIN") - should we read from resource type?
  - **CRITICAL:** Determine where `siteplanId` comes from - do we need a UI page to select this?

- **[cmd/iris-server/iris-server.go](../cmd/iris-server/iris-server.go)**
  - Implement periodic check of MCP resources to update status and resource info in database

### 1.3 System Information
- **[internal/handlers/system.go](../internal/handlers/system.go)**
  - Read version from build-time variable or config
  - Return version information in response

---

## 2. API Routes Comparison

### 2.1 Routes Unchanged
| Route | Method | Python | Go | Status |
|-------|--------|--------|-----|--------|
| `/tracker/` | GET | ✅ | ✅ | Implemented |
| `/resources/` | GET | ✅ | ✅ | Implemented |
| `/system/status` | GET | ✅ | ✅ | Implemented |
| `/system/version` | GET | ✅ | ✅ | Partially implemented |
| `/mcp/operations` | GET | ✅ | ✅ | Scaffolded |
| `/mcp/operations/enable` | POST | ✅ | ✅ | Scaffolded |
| `/mcp/operations/disable` | POST | ✅ | ✅ | Scaffolded |
| `/mcp/start` | POST | ✅ | ✅ | Implemented |

### 2.2 Routes Changed

#### Gateway/Webhook Routes
- **Python:** `/gateway/data` (POST)
- **Go:** `/chirpstackGateway/data` (POST)
- **Reason:** More explicit naming to differentiate future tracker types (TETRA, etc.)

#### Tracker Assignment Routes
- **Python:** `/tracker/{instance_id}` (POST)
  - Body: `TrackerUpdateModel` with resource field
- **Go:** `/tracker/assign/{tracker_id}/{resource_id}` (POST)
  - Resource ID now in URL path
  - Support for unassignment via empty resource_id
  - Uses UUID instead of integer IDs
- **Reason:** More RESTful design, clearer intent, supports resource unassignment

### 2.3 New Routes in Go
| Route | Method | Purpose |
|-------|--------|---------|
| `/tracker/rename/{tracker_id}` | POST | Rename tracker with new name in body |
| `/mcp/config` | GET | Retrieve current MCP configuration |

### 2.4 Routes Not Yet Implemented in Go
- None (all Python routes have Go equivalents, though some are scaffolded)

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
- **Go:** Separate junction table `tracker_resources` with `tracker_id` and `resource_id`
  - One-to-many relationship enforced via UNIQUE constraint on `tracker_id`
  - Supports ON CONFLICT for upsert operations
- **Benefit:** Go version uses cleaner normalized schema, easier to extend to many-to-many if needed

### 3.2 New Tables in Go
- `mcp_config` - Stores MCP server configuration (enabled, url, api_key)
- `tracker_resources` - Junction table for tracker-resource assignments (replaces Python's direct foreign key)

### 3.3 Removed Fields
- **Python `operations` table:** Not yet implemented in Go (TODO: determine if needed)
  - Fields: `id`, `uid`, `title`, `active`, `archived`, `selected`
  - **Status:** Functionality questioned in TODO comment

---

## 4. Functional Differences

### 4.1 Architecture & Code Structure

#### Repository Pattern
- **Python:** CRUD functions directly in `tracking/db/crud.py`
- **Go:** Clean repository layer in `internal/repository/` with interface abstraction
- **Benefit:** Better separation of concerns, easier testing

#### Tracker Abstraction
- **Python:** Concrete `Tracker` class with SQLAlchemy ORM
- **Go:** 
  - `Tracker` interface with `GetPosition()`, `GetLastUpdate()`, `GetBattery()`
  - `BaseTracker` struct with common fields
  - Type-specific trackers extend BaseTracker
- **Benefit:** True polymorphism, easier to add new tracker types

#### Configuration Management
- **Python:** `Settings` class loaded via Pydantic settings
- **Go:** `Config` struct loaded via `caarlos0/env` with structured sub-configs
- **Benefit:** More granular organization (MCPConfig, SQLConfig, SentryConfig)

### 4.2 Chirpstack Message Handling

#### Message Parsing
- **Python:** 
  - Uses Pydantic models with discriminated unions
  - Type checking via `isinstance()`
  - Separate model classes: `ChirpstackPayloadBatteryMessage`, `ChirpstackPayloadLatitudeMessage`, etc.
- **Go:**
  - Uses protobuf `structpb.Struct` from official Chirpstack API
  - `.AsMap()` conversion to `map[string]interface{}`
  - Type assertions for dynamic JSON-like structures
- **Benefit:** Go uses official Chirpstack protobuf definitions for better compatibility

#### Update Logic
- **Python:** Compares message timestamp against `lastUpdated` before applying updates
- **Go:** Currently missing timestamp comparison (TODO in implementation)
- **Impact:** Go version may need to add timestamp validation

### 4.3 MCP Integration

#### Position Mirroring
- **Python:** 
  - Background task via `get_mcp_data()` scheduled function
  - Updates resources from MCP tableau
- **Go:**
  - Direct HTTP client in `internal/mcp_control/positionMirror.go`
  - `SyncTrackerPositionToMCP()` for pushing tracker positions to markers
  - Stores marker ID mapping in database
  - sync from mcp to go still missing
- **Benefit:** When implemented, Go version supports bidirectional sync, Python only pulls resources

#### Operations Management
- **Python:** 
  - Fetches operations from MCP on startup
  - Stores in database with `selected` flag
  - Filters resources by selected operation
- **Go:** 
  - Operations functionality scaffolded but questioned in TODO
  - **DECISION NEEDED:** May simplify to just marker position sync
- **Impact:** Go version may reduce complexity by eliminating operation selection


### 4.4 Error Handling

#### Database Errors
- **Python:** SQLAlchemy exceptions
- **Go:** 
  - pgx errors with `pgconn.PgError` type assertions
  - Explicit error code checking (e.g., `23505` for unique violations)
- **Benefit:** Go has more explicit PostgreSQL error code handling

### 4.5 Concurrency

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

---

## 6. Features Left Out

### 6.1 MCP Operations Table
- **Status:** Scaffolded but not implemented
- **Reason:** Functionality questioned - may only need marker position sync
- **Python Feature:** Full operation selection with `active`/`selected` flags
- **Impact:** Simpler Go implementation if removed

### 6.2 Background Resource Sync
- **Status:** TODO in main.go
- **Python Feature:** Scheduled `get_mcp_data()` task pulls resources from MCP
- **Impact:** Go version needs goroutine with periodic execution

### 6.3 TETRA Implementation
- **Status:** Models and database schema ready, no implementation
- **Python Feature:** N/A (not in Python version either)
- **Impact:** Planned feature for both versions

### 6.4 FastAPI Automatic Documentation
- **Status:** No equivalent in Go/Gin
- **Python Feature:** Auto-generated OpenAPI/Swagger docs
- **Impact:** May need manual API documentation or Swagger annotations


---

## 8. Migration Considerations

### 8.1 Data Migration
- **Database:** Schema changes require migration script
  - Convert integer IDs to UUIDs for resources
  - Split trackers into base + type-specific tables
  - Migrate `lastUpdated` unix timestamps to PostgreSQL timestamps
- **Approach:** Write migration script or fresh database initialization

### 8.2 API Compatibility
- **Frontend Changes Needed:**
  - Update Chirpstack webhook URL: `/gateway/data` → `/chirpstackGateway/data`
  - Update tracker assignment endpoint: `/tracker/{id}` → `/tracker/assign/{tracker_id}/{resource_id}`
  - Handle UUID strings instead of integers for IDs
  - New tracker rename endpoint available

### 8.3 Configuration Changes
- **Environment Variables:**
  - MCP configuration now includes `MCP_ENABLED`, `MCP_SERVER_URL`, `MCP_API_KEY`
  - SQL configuration more granular: `SQL_HOST`, `SQL_PORT`, `SQL_USER`, `SQL_PASSWORD`, `SQL_DB_NAME`
  - New: `SENTRY_DSN`, `SENTRY_ENVIRONMENT`

---

## 9. Next Steps

### 9.1 Critical Path
1. ✅ Complete Chirpstack message parsing (battery, lat/long) - **DONE**
2. ⚠️ **FIX CRITICAL BUG:** Implement `CreateChirpstackTracker()` - currently just returns nil!
3. ⚠️ **FIX CRITICAL BUG:** Fix `UpdateTracker()` - wrong field names, wrong SQL columns, inverted logic
4. ✅ Implement Chirpstack webhook handler structure - **DONE** (but broken due to bugs above)
5. ✅ Implement tracker rename functionality - **DONE**
6. ✅ Implement MCP position sync (POST/PUT markers) - **DONE**
7. ⚠️ Decide on MCP operations functionality (keep or simplify)
8. 🔲 Save marker ID to resource in database (TODO in positionMirror.go)
9. 🔲 Resolve siteplanId requirement for MCP markers
10. 🔲 Add background goroutine for periodic MCP resource sync
11. 🔲 Complete system version endpoint
12. 🔲 Testing with real Chirpstack webhooks and MCP integration (cannot test until bugs fixed)

### 9.2 Future Enhancements
- TETRA tracker implementation
- API documentation (Swagger/OpenAPI)
- Database migration scripts
- Unit and integration tests
- Deployment configuration (Docker, docker-compose)

---

## 10. Open Questions

1. **MCP Operations:** Do we need the full operation enable/disable functionality, or just marker position sync?
2. **Siteplan ID:** Where does `siteplanId` come from for MCP markers? Need UI page for selection?
3. **Marker Icons:** Should icon type be determined from resource type, or remain hardcoded?
4. **Version Info:** Should version be build-time variable, config file, or git tag?
5. **Background Tasks:** What interval for periodic MCP resource checks? Use cron-like scheduler or simple ticker?

---

## Document Status
- **Created:** January 28, 2026
- **Last Updated:** January 28, 2026 (Verified and updated with critical bug discoveries)
- **Maintained By:** Development team
- **Review Schedule:** Update after major milestones
