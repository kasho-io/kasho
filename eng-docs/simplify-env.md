# Environment Variable Simplification Plan

This document outlines the plan to simplify environment variable configuration in Kasho by moving from decomposed variables to URL-based configuration.

## Current State

Currently, Kasho uses decomposed environment variables:
- `PRIMARY_DATABASE_HOST`, `PRIMARY_DATABASE_PORT`, `PRIMARY_DATABASE_KASHO_USER`, etc.
- `REPLICA_DATABASE_HOST`, `REPLICA_DATABASE_PORT`, `REPLICA_DATABASE_KASHO_USER`, etc.
- `CHANGE_STREAM_HOST`, `CHANGE_STREAM_PORT`

## Target State

Simplified URL-based configuration:
- **pg-change-stream**: `KV_URL`, `PRIMARY_DATABASE_URL`
- **pg-translicator**: `CHANGE_STREAM_SERVICE_ADDR`, `REPLICA_DATABASE_URL`

Where:
- `PRIMARY_DATABASE_URL`: `postgresql://kasho_user:password@host:port/dbname?sslmode=disable`
- `REPLICA_DATABASE_URL`: `postgresql://kasho_user:password@host:port/dbname?sslmode=disable`
- `CHANGE_STREAM_SERVICE_ADDR`: `pg-change-stream:50051` (host:port format)

## Implementation Plan

### Phase 1: Update Service Code to Accept URL Variables

#### pg-change-stream Service
- [x] Modify configuration loading to accept `PRIMARY_DATABASE_URL`
- [x] Parse the URL to extract connection parameters
- [x] Maintain backward compatibility with decomposed variables
- [x] Update connection string building logic
- [x] Test with both URL and decomposed variables

#### pg-translicator Service
- [x] Modify configuration loading to accept `REPLICA_DATABASE_URL`
- [x] Modify configuration loading to accept `CHANGE_STREAM_SERVICE_ADDR`
- [x] Parse database URL to extract connection parameters
- [x] Parse service address to extract host and port
- [x] Maintain backward compatibility with decomposed variables
- [x] Test with both URL and decomposed variables

### Phase 2: Install and Configure trurl

#### Add trurl to Docker Images
- [x] Install trurl in Dockerfile (development and production stages)
- [x] Test trurl installation and basic functionality
- [x] Document trurl version for reproducibility

#### Create URL Parsing Wrapper Script
- [x] Create shell script that uses trurl to parse database URLs
- [x] Extract all PostgreSQL URL components using trurl
- [x] Handle query parameters (sslmode, etc.)
- [x] Output as shell-compatible export statements
- [x] Add error handling for malformed URLs

Example usage:
```bash
# Using trurl to extract components
$ export PRIMARY_DATABASE_URL="postgresql://user:pass@host:5432/db?sslmode=disable"

# Extract individual components
$ export DB_USER=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{user}')
$ export DB_PASSWORD=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{password}')
$ export DB_HOST=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{host}')
$ export DB_PORT=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{port}')
$ export DB_NAME=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{path}' | sed 's|^/||')
$ export DB_SSLMODE=$(trurl --url "$PRIMARY_DATABASE_URL" --get '{query:sslmode}')
```

### Phase 3: Update Environment Configuration

#### Update .env Files
- [x] Update demo/.env to use URL variables
- [x] Update development/.env to use URL variables
- [x] Remove decomposed variables from .env files
- [x] Add clear comments explaining the URL format

#### Update Docker Compose Files
- [x] Update demo/docker-compose.yml environment sections
- [x] Update development/docker-compose.yml environment sections
- [x] Ensure KV_URL is properly set (currently missing)

#### Update SQL Templates Bridge
- [x] Create wrapper script for env-template that:
  - Uses trurl to parse URL environment variables
  - Sets decomposed variables for templates
  - Handles both PRIMARY_DATABASE_URL and REPLICA_DATABASE_URL
  - Calls env-template with decomposed variables
- [x] Update docker-compose to use wrapper script
- [x] Test SQL template generation with various URL formats

### Phase 4: Update Scripts and Tools

#### Update Bash Scripts
- [x] Update environments/demo/run.sh
- [x] Update environments/development/run.sh
- [x] Update environments/demo/test.sh
- [x] Update any other scripts using database variables

#### Update pg-bootstrap-sync
- [x] Review if any changes needed (currently uses CLI flags)
- [x] Update documentation if needed

### Phase 5: Documentation and Help

#### Create Help Script
- [x] Create /app/kasho-help script for Docker CMD
- [x] Document all environment variables for each service
- [x] Include examples of URL formats
- [x] Add troubleshooting tips

#### Update Documentation
- [x] Update README.md with new environment variables
- [x] Update deployment documentation
- [x] Create migration guide for existing users
- [x] Update docker-compose examples

### Phase 6: Testing and Validation

#### Integration Testing
- [x] Test demo environment with new variables
- [x] Test development environment with new variables
- [x] Test backward compatibility mode
- [x] Test SQL template generation
- [x] Test all services can connect properly

#### Edge Cases
- [x] Test with special characters in passwords
- [x] Test with non-standard ports
- [x] Test with various sslmode settings
- [x] Test missing optional parameters

### Phase 7: Cleanup

#### Remove Backward Compatibility
- [x] Remove support for decomposed variables from services
- [x] Update all documentation to only show URL format
- [x] Archive this migration plan

## Migration Strategy

To ensure smooth transition:
1. Services will temporarily support both URL and decomposed variables
2. URL variables take precedence if both are present
3. Clear deprecation warnings for decomposed variables
4. Provide migration period before removing old format

## Benefits

1. **Simpler Configuration**: Fewer environment variables to manage
2. **Standard Format**: PostgreSQL URLs are industry standard
3. **Easier Deployment**: Single variable per database connection
4. **Better Portability**: URLs can be easily copied between environments
5. **Cleaner Code**: Less boilerplate for connection handling

## Risks and Mitigations

1. **Risk**: Breaking existing deployments
   - **Mitigation**: Backward compatibility during transition

2. **Risk**: Complex passwords with special characters
   - **Mitigation**: Proper URL encoding/decoding in parser

3. **Risk**: SQL templates need decomposed variables
   - **Mitigation**: Parser utility bridges the gap

4. **Risk**: Loss of granular control
   - **Mitigation**: Query parameters provide flexibility

## Timeline

- Phase 1-2: Core implementation (1-2 days)
- Phase 3-4: Configuration updates (1 day)
- Phase 5-6: Documentation and testing (1 day)
- Phase 7: Cleanup after validation period (1 week later)