> I need to design a licensing service for some SaaS software that runs on the customer’s own infrastructure. What are some good approaches? Let’s just discuss — please don’t show me any code.

There are several solid approaches for SaaS licensing when your software runs on customer infrastructure. Here are the main patterns to consider:

**License Key Validation**
The simplest approach uses cryptographically signed license keys that your software validates locally. You generate keys containing license details (expiration, feature flags, user limits) and sign them with your private key. The customer software validates using your public key. This works offline but makes it harder to enforce real-time compliance or gather usage analytics.

**Phone-Home Licensing**
Your software periodically contacts your licensing server to validate and renew its license. This gives you real-time control and usage visibility, but requires the customer's infrastructure to have internet access. You can make it graceful by allowing temporary offline operation with cached licenses.

**Hardware Fingerprinting**
Tie licenses to specific hardware characteristics (CPU ID, MAC address, disk serial numbers) to prevent casual license sharing. This works well but can be fragile when customers upgrade hardware or run in virtualized environments.

**Floating/Concurrent Licenses**
Instead of node-locked licenses, track concurrent usage across the customer's infrastructure. Users "check out" licenses when using the software and release them when done. This maximizes license utilization but requires a license server component.

**Feature-Based Licensing**
Structure licenses around specific features or capacity limits rather than just time-based expiration. For example, "up to 1000 transactions/day" or "advanced analytics module enabled." This aligns pricing with value delivered.

**Hybrid On-Premise License Server**
Provide a lightweight license server that customers deploy in their own infrastructure. It handles local license distribution while occasionally syncing with your central system. This balances control with customer data sovereignty concerns.

What type of software are you licensing, and do you have any constraints around internet connectivity or customer data policies that might influence the approach?

> I like the idea of phone home with caching and support for different features. I’d also like to be able to extend the license format later to support additional constraints, features, or use cases.

That's a solid combination. Phone-home with caching gives you the best of both worlds - real-time control when connected, graceful degradation when offline. Here are some key design considerations for your approach:

**License Format Extensibility**
Design your license structure as a versioned schema from day one. JSON Web Tokens (JWT) work well here since they're self-contained, cryptographically signed, and support custom claims. You can include a schema version field and design your validation logic to handle multiple versions gracefully. Alternatively, a custom binary format with version headers gives you more control over size and parsing.

**Caching Strategy**
Consider how long licenses should remain valid offline. You might use different cache durations for different license types - maybe enterprise customers get longer offline periods than trial users. Implement graceful degradation where some features might be disabled after extended offline periods, but core functionality continues.

**Feature Flag Architecture**
Structure your feature system to be additive. Instead of just boolean flags, consider using capability objects that can express limits, quotas, or configuration parameters. For example, rather than just "analytics_enabled: true", you might have "analytics: {enabled: true, retention_days: 90, max_queries_per_hour: 1000}".

**Constraint Types**
Think about different constraint patterns you might need: time-based (expiration, usage windows), usage-based (API calls, data volume, concurrent users), geography-based (deployment regions), or organizational (department restrictions, user groups).

**Phone-Home Protocol**
Design this to be lightweight and batchable. Your software might cache multiple events and send them together, rather than making a call for every license check. Consider what happens if the licensing service is temporarily unavailable - you'll want exponential backoff and circuit breaker patterns.

How frequently are you thinking of having the software check in, and what kinds of features are you planning to gate initially?

> I’m planning to use workload pricing. My product is something that works on databases, providing both replication as well as transformation of columns based on predefined rules. It also supports DDL replication when the structure changes. So there will be usage metrics about rows replicated, columns transformed, and structure changes.

That's a perfect fit for workload-based licensing with phone-home architecture. Database replication and transformation tools naturally generate measurable usage metrics, and your customers will likely appreciate paying for what they actually use rather than upfront capacity estimates.

**Usage Metering Design**
You'll want to batch and buffer your usage metrics locally before sending them to avoid impacting database performance. Consider tracking metrics like rows replicated per table, transformation operations by type/complexity, and DDL events. You might also want to capture metadata like source/target database types, table sizes, and transformation rule complexity to support tiered pricing later.

**Pricing Model Flexibility**
Your license format should support different pricing dimensions that you can mix and match. For example, you might price transformation differently than straight replication, or charge more for complex transformations versus simple column mapping. DDL replication might be priced per schema change or as a premium feature add-on.

**Usage Aggregation Windows**
Think about how you'll handle usage reporting - daily, hourly, or real-time. Database workloads can be bursty, so you might want to support different aggregation periods. Some customers might prefer monthly billing cycles, while others want real-time usage visibility.

**Quota Management**
Your caching strategy becomes crucial here. When offline, you'll need to decide whether to allow usage to continue accumulating (with later reconciliation) or enforce last-known quotas. You might implement soft limits with grace periods versus hard cutoffs.

**Multi-Environment Considerations**
Database environments often span dev/test/prod with different usage patterns. Your licensing might need to handle different rate limits or pricing for non-production environments.

Are you thinking of pricing all operations equally, or do you anticipate different rates for replication versus transformation workloads?

---

## Kasho Licensing Implementation Plan

### JWT License Format

Simple format supporting expiration only (usage tracking to be added later):

```json
{
  "iss": "kasho.io",
  "sub": "customer-uuid",
  "iat": 1704067200,
  "exp": 1735689600,
  "kasho": {
    "version": 1,
    "customer_name": "Acme Corp"
  }
}
```

### License File Location

The license file will be at `/app/config/license.jwt`, in the same directory as `transforms.yml`. This keeps all configuration files in one consistent location.

### Licensing Service

**Service Name:** `licensing` (located at `services/licensing/`)

**Responsibilities:**
- License validation
- License refresh from home (future)
- Usage metrics reporting (future)

**gRPC Service Definition:**

```proto
syntax = "proto3";

package kasho.license;

option go_package = "kasho/proto/license;license";

service License {
  // ValidateLicense checks if the current license is valid
  rpc ValidateLicense(ValidateLicenseRequest) returns (ValidateLicenseResponse) {}
  
  // GetLicenseInfo returns current license details
  rpc GetLicenseInfo(GetLicenseInfoRequest) returns (GetLicenseInfoResponse) {}
}

message ValidateLicenseRequest {}

message ValidateLicenseResponse {
  bool valid = 1;
  string reason = 2; // empty if valid, error message if invalid
  int64 expires_at = 3; // Unix timestamp
}

message GetLicenseInfoRequest {}

message GetLicenseInfoResponse {
  string customer_id = 1;
  string customer_name = 2;
  int64 issued_at = 3;
  int64 expires_at = 4;
  bool valid = 5;
}
```

### Developer Experience

**Developer License Utility** (`tools/dev-license/`):
- Generates valid development licenses with configurable expiration
- Can watch and auto-regenerate licenses before expiration
- Embeds the same public key the services will use
- Usage: `dev-license --days 30 --output /app/config/license.jwt`

### Integration Strategy

**Common Pattern:**
1. Import shared `pkg/license` client library
2. Initialize license client with gRPC connection
3. Check license validity on startup
4. Handle three cases:
   - Valid license: proceed normally
   - Invalid/expired: exit with error
   - License service unavailable: allow with warning (development mode)

**Service-Specific Integration:**

**pg-change-stream:**
- Check on startup before accepting connections
- Periodic revalidation (hourly)
- Log warnings but continue operation if license expires during runtime

**pg-translicator:**
- Check on startup
- Cache validation for 5 minutes
- Validate before processing transformations

**pg-bootstrap-sync:**
- Check before processing any dump file
- Exit immediately if invalid

### Testing Strategy

**1. Licensing Service Tests** (`services/licensing/internal/`):
- Unit tests for JWT parsing and validation
- Tests for expired licenses, invalid signatures, malformed JWTs
- gRPC endpoint tests with mock licenses
- Integration tests with file system operations

**2. Integration Tests for Existing Services:**

**pg-change-stream:**
- Test startup with valid license
- Test startup with expired license (should fail)
- Test behavior when license expires during operation
- Test when licensing service is unavailable

**pg-translicator:**
- Similar startup tests
- Test that transformations are blocked with invalid license
- Test license validation caching

**pg-bootstrap-sync:**
- Test that it refuses to process dumps without valid license
- Test successful operation with valid license

**3. End-to-End Tests:**
- Docker compose environment with all services
- Test license expiration scenarios
- Test license refresh functionality (when implemented)

**4. Test Utilities:**
- Shared test helper package (`pkg/license/testutil/`)
- Functions to generate test licenses with various states
- Mock licensing service for unit tests

### Versioning Strategy

**Shared Version Package** (`pkg/version/version.go`):
```go
package version

import (
    "fmt"
    "strconv"
    "strings"
)

var (
    // Version is the full semver version (set at build time)
    Version = "1.0.0"
    
    // GitCommit is the git commit hash (set at build time)
    GitCommit = ""
    
    // BuildDate is the build timestamp (set at build time)
    BuildDate = ""
)

// MajorVersion returns the major version number
func MajorVersion() int {
    parts := strings.Split(Version, ".")
    if len(parts) > 0 {
        major, _ := strconv.Atoi(parts[0])
        return major
    }
    return 1
}

// MinorVersion returns the minor version number
func MinorVersion() int {
    parts := strings.Split(Version, ".")
    if len(parts) > 1 {
        minor, _ := strconv.Atoi(parts[1])
        return minor
    }
    return 0
}

// PatchVersion returns the patch version number
func PatchVersion() int {
    parts := strings.Split(Version, ".")
    if len(parts) > 2 {
        patch, _ := strconv.Atoi(parts[2])
        return patch
    }
    return 0
}

// Info returns structured version information
func Info() VersionInfo {
    return VersionInfo{
        Version:   Version,
        Major:     MajorVersion(),
        Minor:     MinorVersion(),
        Patch:     PatchVersion(),
        GitCommit: GitCommit,
        BuildDate: BuildDate,
    }
}

type VersionInfo struct {
    Version   string
    Major     int
    Minor     int
    Patch     int
    GitCommit string
    BuildDate string
}
```

**Git Tagging and Versioning Strategy**:

Git tags should use the `v` prefix (e.g., `v1.2.3`) for Go module compatibility, but the application version should be pure semver (e.g., `1.2.3`).

Release process:
```bash
# Tag a new release
git tag v1.2.3
git push origin v1.2.3

# Development builds between releases will show:
# v1.2.3-7-gabc1234 (7 commits after tag)
# v1.2.3-7-gabc1234-dirty (with uncommitted changes)
```

**Taskfile.yml Updates**:
```yaml
version: '3'

vars:
  GIT_TAG:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0"
  VERSION:
    sh: echo "{{.GIT_TAG}}" | sed 's/^v//'  # Strip v prefix for semver compliance
  COMMIT:
    sh: git rev-parse HEAD 2>/dev/null || echo "unknown"
  DATE:
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"
  LDFLAGS: >-
    -X kasho/pkg/version.Version={{.VERSION}}
    -X kasho/pkg/version.GitCommit={{.COMMIT}}
    -X kasho/pkg/version.BuildDate={{.DATE}}

tasks:
  build:
    desc: Build the kasho Docker image for development
    deps: [build:base]
    cmds:
      - docker build -t kasho --target development --build-arg LDFLAGS="{{.LDFLAGS}}" .

  build:prod:
    desc: Build the kasho Docker image for production
    deps: [build:base]
    cmds:
      - docker build -t kasho:prod --target production --build-arg LDFLAGS="{{.LDFLAGS}}" .

  # New task for version info
  version:
    desc: Display version information
    cmds:
      - echo "Version: {{.VERSION}}"
      - echo "Commit: {{.COMMIT}}"
      - echo "Build Date: {{.DATE}}"
```

**Dockerfile Updates**:
```dockerfile
# Accept LDFLAGS as build argument
ARG LDFLAGS=""

# Use LDFLAGS when building services
RUN go build -ldflags "${LDFLAGS}" -o /app/service ./cmd/server
```

**transforms.yml Changes**:
- Change `version: v1` to `major_version: 1`
- Update validation in `transform/config.go` to check against `version.MajorVersion()`
- Add clear error messages when major versions don't match

**License JWT Format Update**:
```json
{
  "iss": "kasho.io",
  "sub": "customer-uuid",
  "iat": 1704067200,
  "exp": 1735689600,
  "kasho": {
    "major_version": 1,  // Changed from "version" to be consistent
    "customer_name": "Acme Corp"
  }
}
```

**Version Compatibility Checks**:
- License service validates that license `kasho.major_version` <= current Kasho major version
- Transform config validates that `major_version` == current Kasho major version
- Services report version in logs, metrics, and gRPC metadata

### Implementation Order

1. Create shared version package
2. Update build scripts to inject version information (Taskfile.yml, Dockerfile)
3. Update transforms.yml format and validation code
4. Create proto definition and generate code
5. Implement licensing service with file-based license reading
6. Create shared license client package
7. Build dev-license utility
8. Integrate with pg-change-stream
9. Integrate with pg-translicator
10. Integrate with pg-bootstrap-sync
11. Add comprehensive tests
12. Update Docker compose files and documentation