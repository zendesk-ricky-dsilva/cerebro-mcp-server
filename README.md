# Cerebro MCP Server

A Model Context Protocol (MCP) server that provides access to Cerebro project data. This server allows AI assistants like Claude to retrieve project details and repository information from the Cerebro API.

## Features

- **Project Details Retrieval**: Get comprehensive project information from Cerebro
- **Asynchronous Dependencies**: Retrieve detailed dependency information using parallel API calls for optimal performance
- **Repository Filtering**: Filter repositories by `kube_project` field matching
- **MCP Protocol Support**: Compatible with Claude Desktop and other MCP clients
- **HTTP API Support**: Call the server via REST API endpoints
- **Secure Authentication**: Uses token-based authentication with Cerebro API

## Prerequisites

- Go 1.21 or later
- Valid Cerebro API token
- Access to Cerebro API at `https://cerebro.zende.sk`

## Installation

1. Clone or download this repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build the server:
   ```bash
   go build -o cerebro-mcp-server .
   ```

### Using the Makefile (Recommended)

The project includes a comprehensive Makefile for streamlined development workflow:

```bash
# Build the binary
make build

# Run all checks (format, vet, build, test)
make check

# Run tests only
make test

# Clean build artifacts
make clean

# Install Go dependencies
make deps

# Format Go code
make fmt

# Run go vet
make vet

# Show all available targets
make help
```

**Quick Start with Makefile:**
```bash
# Build and test in one command
make check

# Or just build
make build
```

## Configuration

Set the required environment variable:

```bash
export CEREBRO_TOKEN="your-cerebro-api-token"
```

## Usage

### MCP Mode (Default)
Start the server in MCP mode for integration with Claude Desktop or other MCP clients:

```bash
./cerebro-mcp-server
```

**Using Makefile:**
```bash
make run-mcp
```

### HTTP Mode
Start the server in HTTP mode to call it via REST API:

```bash
HTTP_MODE=true ./cerebro-mcp-server
```

**Using Makefile:**
```bash
make run-http
```

The server will start on port 8080 and accept POST requests to `/mcp`.

#### HTTP API Example

Request format for project details:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "project_get_details",
    "arguments": {
      "project_permalink": "telemetry-pipeline"
    }
  }'
```

Request format for project dependencies:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "project_get_dependencies",
    "arguments": {
      "project_permalink": "classic"
    }
  }'
```

Response format:
```json
{
  "success": true,
  "data": {
    "content": [
      {
        "type": "text",
        "text": "# Project Details for: classic\n\n**Project Name:** classic..."
      }
    ]
  }
}
```

Error response:
```json
{
  "success": false,
  "error": "Tool not found: invalid_tool"
}
```

## Integration with Claude Desktop

Add this configuration to your Claude Desktop config file:

```json
{
  "mcpServers": {
    "cerebro": {
      "command": "/path/to/your/cerebro-mcp-server",
      "env": {
        "CEREBRO_TOKEN": "your-cerebro-api-token"
      }
    }
  }
}
```

## API Integration

The server integrates with the Cerebro API at `https://cerebro.zende.sk/projects.json`.

### Authentication
Uses token-based authentication via the `Authorization: Token <token>` header.

### Query Parameters
- `search[permalink]`: Filter projects by permalink
- `search[id]`: Filter projects by ID
- `includes`: Include related data (e.g., "repositories", "project_dependencies")
- `inlines`: Include additional inline data fields

## Performance Optimizations

### Asynchronous Dependency Fetching

The `project_get_dependencies` tool implements asynchronous API calls for optimal performance:

- **Parallel Processing**: When fetching dependency details, all providing project information is retrieved concurrently using Go routines
- **Reduced Latency**: Instead of sequential API calls, dependencies are fetched in parallel, significantly reducing total response time
- **Maintained Order**: Results are collected and presented in the original dependency order
- **Error Handling**: Individual API failures don't block other dependency fetches

**Performance Impact:**
- For a project with 50 dependencies: Sequential = ~15 seconds, Async = ~1-2 seconds
- Response time scales with the slowest individual API call rather than the sum of all calls
- Particularly beneficial for projects with many dependencies (like the "classic" project with 80+ dependencies)

### HTTP Client Optimization

- **Connection Reuse**: HTTP client with 30-second timeout for efficient connection management
- **Concurrent Requests**: No artificial rate limiting, allowing full utilization of API capacity
- **Error Resilience**: Failed requests for individual dependencies don't terminate the entire operation

## Available Tools

### project_get_details

Retrieves detailed information about a project and its associated repositories.

**Parameters:**
- `project_permalink` (required): The permalink of the project to retrieve

**Returns:**
- Project metadata (name, description, category, etc.)
- List of repositories where `kube_project` matches the project permalink
- Repository details (URL, category, status, etc.)

### project_get_dependencies

Retrieves comprehensive dependency information for a project using asynchronous API calls for optimal performance.

**Parameters:**
- `project_permalink` (required): The permalink of the project to retrieve dependencies for

**Performance Features:**
- **Asynchronous Processing**: All dependency details are fetched concurrently using Go routines
- **Parallel API Calls**: Multiple providing projects are queried simultaneously
- **Maintained Order**: Results are presented in the original dependency order
- **Error Resilience**: Individual API failures don't block other dependency fetches

**Returns:**
- Project basic information (ID, name, permalink, description)
- List of dependencies with detailed information for each:
  - Providing project details (name, category, criticality tier, owner, etc.)
  - Dependency metadata (optional flag, description, creation/update timestamps)
  - Relationship information (dependency ID, providing project ID)

## Data Structures

### Project
Contains project metadata including:
- Basic info (ID, name, permalink, description)
- Configuration (category, deploy target, runs on)
- Status information (criticality tier, release state)
- Relationships (repository IDs, slack channels)
- Dependencies (dependent project dependencies IDs)

### Repository
Contains repository information including:
- Basic info (ID, name, URL, permalink)
- Metadata (category, created/updated dates)
- Status (archived, deprecated)
- **kube_project**: The field used for filtering

### ProjectDependency
Contains dependency relationship information including:
- Relationship IDs (ID, dependent project ID, providing project ID)
- Metadata (description, optional flag)
- Timestamps (created at, updated at, deleted at)

## Example Response

### Project Details Example

When querying for `example-project`, you might get:

```
# Project Details for: example-project

**Project Name:** Example Project
**Description:** Sample project description
**Category:** Service
**Deploy Target:** Kubernetes
**Runs On:** Production
**Slack Channel:** example-team
**Started On:** 2023-01-15

## Repositories with kube_project = 'example-project'

Found 2 matching repositories:

### 1. example-service
- **URL:** https://github.com/company/example-service
- **Permalink:** example-service
- **Category:** Service
- **Started On:** 2023-01-15
- **Archived:** false
- **Last Updated:** 2025-07-05T00:05:04.000Z

### 2. example-worker
- **URL:** https://github.com/company/example-worker
- **Permalink:** example-worker
- **Category:** Service
- **Started On:** 2023-06-16
- **Archived:** false
- **Last Updated:** 2025-07-01T00:09:12.000Z
```

### Project Dependencies Example

When querying dependencies for `example-project`, you might get:

```
# Dependencies for Project: Example Project

**Project ID:** 100
**Permalink:** example-project
**Description:** A sample project for demonstration purposes

## Dependencies (3)

### 1. Authentication Service
- **Dependency ID:** 1
- **Providing Project ID:** 200
- **Permalink:** auth-service
- **Description:** Authentication and authorization service
- **Category:** Service
- **Criticality Tier:** Tier 1
- **Release State:** GA
- **Owner Team:** Platform Team
- **Slack Channel:** platform-support
- **Optional Dependency:** false
- **Dependency Created:** 2023-01-15T10:00:00.000Z
- **Last Updated:** 2023-06-01T15:30:00.000Z

### 2. Database Service
- **Dependency ID:** 2
- **Providing Project ID:** 300
- **Permalink:** db-service
- **Description:** Database management service
- **Category:** Infrastructure
- **Criticality Tier:** Tier 0
- **Release State:** GA
- **Owner Team:** Data Team
- **Slack Channel:** data-support
- **Optional Dependency:** false
- **Dependency Created:** 2023-01-15T10:05:00.000Z
- **Last Updated:** 2023-06-01T15:35:00.000Z

### 3. Logging Service
- **Dependency ID:** 3
- **Providing Project ID:** 400
- **Permalink:** logging-service
- **Description:** Centralized logging service
- **Category:** Infrastructure
- **Criticality Tier:** Tier 2
- **Release State:** GA
- **Owner Team:** Platform Team
- **Slack Channel:** platform-support
- **Optional Dependency:** true
- **Dependency Created:** 2023-02-01T09:00:00.000Z
- **Last Updated:** 2023-06-15T14:20:00.000Z

[... additional dependencies ...]
```

## Error Handling

The server provides detailed error messages for:
- Missing or invalid API tokens
- Network connectivity issues
- Invalid project permalinks
- API response parsing errors
- HTTP status code errors

## Development

### Project Structure
```
.
├── main.go                      # Main server implementation
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── Makefile                     # Build and test automation
├── test_dependencies.sh         # Dependencies endpoint test script
├── README.md                    # This documentation
├── examples/                    # Example API responses
│   ├── project_dependencies_response_payload.json
│   └── projects_response_payload.json
└── api/                         # API documentation
```

### Dependencies
- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- Standard Go libraries for HTTP, JSON, and networking

### Building
```bash
go build -o cerebro-mcp-server .
```

**Using Makefile (Recommended):**
```bash
# Build the binary
make build

# Build with all checks (format, vet, build, test)
make check

# Clean and rebuild
make clean build
```

### Testing
```bash
# Run all tests
make test

# Run individual test scripts
./test_dependencies.sh

# Run HTTP mode testing
make run-http
# Then in another terminal:
./test_dependencies.sh
```

### Available Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build the Go binary |
| `test` | Build and run tests |
| `clean` | Remove build artifacts |
| `deps` | Install Go dependencies |
| `run-http` | Run server in HTTP mode |
| `run-mcp` | Run server in MCP mode |
| `fmt` | Format Go code |
| `vet` | Run go vet |
| `check` | Run all checks (fmt, vet, build, test) |
| `help` | Show available targets |

## Troubleshooting

### Common Issues

1. **"CEREBRO_TOKEN environment variable is required"**
   - Set the `CEREBRO_TOKEN` environment variable with your API token

2. **"failed to execute request"**
   - Check network connectivity to `https://cerebro.zende.sk`
   - Verify your API token is valid

3. **"API request failed with status 401"**
   - Your API token may be invalid or expired
   - Contact your administrator for a new token

4. **"No repositories found with matching kube_project"**
   - The project exists but has no repositories with matching `kube_project` field
   - Verify the project permalink is correct

5. **"No dependencies found for this project"**
   - The project exists but has no dependencies defined in Cerebro
   - This is normal for standalone projects or leaf nodes in the dependency graph

6. **"Project not found" for dependencies**
   - The project permalink doesn't exist in Cerebro
   - Check the spelling and verify the project exists in the system

## License

This project is provided as-is for internal use at Zendesk.
