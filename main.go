package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Constants
const (
	ToolProjectGetDetails      = "project_get_details"
	ToolProjectGetDependencies = "project_get_dependencies"
)

// ProjectServer represents the MCP server
type ProjectServer struct {
	service   *ProjectService
	validator *Validator
}

// createMCPResult creates a standardized MCP result with text content
func createMCPResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// NewProjectServer creates a new Project MCP server
func NewProjectServer(service *ProjectService, validator *Validator) *ProjectServer {
	return &ProjectServer{
		service:   service,
		validator: validator,
	}
}

// SetupMCPServer configures the MCP server with all tools and resources
func (ps *ProjectServer) SetupMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"project-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Add the project_get_details tool
	mcpServer.AddTool(mcp.NewTool("project_get_details",
		mcp.WithDescription("Get details about a project"),
		mcp.WithString("project_permalink",
			mcp.Description("The project permalink to retrieve details for"),
			mcp.Required(),
		),
	), ps.handleGetProjectDetails)

	// Add the project_get_dependencies tool
	mcpServer.AddTool(mcp.NewTool("project_get_dependencies",
		mcp.WithDescription("Get dependency information for a project"),
		mcp.WithString("project_permalink",
			mcp.Description("The project permalink to retrieve dependencies for"),
			mcp.Required(),
		),
	), ps.handleGetProjectDependencies)

	return mcpServer
}

// Tool handlers

func (ps *ProjectServer) handleGetProjectDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Validate and extract project permalink
	projectPermalink, err := ps.validator.ValidateToolArguments(arguments)
	if err != nil {
		return nil, err
	}

	// Get project details using the service
	result, err := ps.service.GetProjectDetails(ctx, projectPermalink)
	if err != nil {
		return nil, err
	}

	return createMCPResult(result.FormattedText), nil
}

func (ps *ProjectServer) handleGetProjectDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Validate and extract project permalink
	projectPermalink, err := ps.validator.ValidateToolArguments(arguments)
	if err != nil {
		return nil, err
	}

	// Get project dependencies using the service
	result, err := ps.service.GetProjectDependencies(ctx, projectPermalink)
	if err != nil {
		return nil, err
	}

	return createMCPResult(result.FormattedText), nil
}

// ServeHTTP implements http.Handler to allow the MCP server to be called via HTTP
func (ps *ProjectServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(HTTPResponse{
			Success: false,
			Error:   "Only POST method is allowed",
		})
		return
	}

	// Parse the request body
	var httpReq HTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(HTTPResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		})
		return
	}

	// Support project_get_details and project_get_dependencies tools
	if httpReq.Tool != ToolProjectGetDetails && httpReq.Tool != ToolProjectGetDependencies {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(HTTPResponse{
			Success: false,
			Error:   fmt.Sprintf("Tool not found: %s", httpReq.Tool),
		})
		return
	}

	// Create a proper MCP CallToolRequest
	mcpRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      httpReq.Tool,
			Arguments: httpReq.Arguments,
		},
	}

	// Call the appropriate tool handler
	var result *mcp.CallToolResult
	var err error

	switch httpReq.Tool {
	case ToolProjectGetDetails:
		result, err = ps.handleGetProjectDetails(context.Background(), mcpRequest)
	case ToolProjectGetDependencies:
		result, err = ps.handleGetProjectDependencies(context.Background(), mcpRequest)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HTTPResponse{
			Success: false,
			Error:   fmt.Sprintf("Tool execution failed: %v", err),
		})
		return
	}

	// Return successful response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HTTPResponse{
		Success: true,
		Data:    result,
	})
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create dependencies
	client := NewCerebroClient(config.CerebroAPIBaseURL, config.CerebroToken)
	validator := NewValidator()
	service := NewProjectService(client, validator)
	projectServer := NewProjectServer(service, validator)

	// Setup MCP server with tools and resources
	mcpServer := projectServer.SetupMCPServer()

	// Check if HTTP_MODE environment variable is set
	if config.HTTPMode {
		startHTTPServer(projectServer, config)
	} else {
		startStdioServer(mcpServer)
	}
}

func startHTTPServer(projectServer *ProjectServer, config *Config) {
	// Start HTTP server
	http.Handle(config.MCPEndpoint, projectServer)
	log.Printf("Project MCP Server starting HTTP mode on %s", config.ServerPort)
	log.Printf("Send POST requests to http://localhost%s%s", config.ServerPort, config.MCPEndpoint)
	log.Printf("Available tools: %s, %s", ToolProjectGetDetails, ToolProjectGetDependencies)
	log.Printf("Example request body: {\"tool\": \"%s\", \"arguments\": {\"project_permalink\": \"your-project\"}}", ToolProjectGetDetails)
	log.Printf("Example dependencies request: {\"tool\": \"%s\", \"arguments\": {\"project_permalink\": \"your-project\"}}", ToolProjectGetDependencies)
	if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func startStdioServer(mcpServer *server.MCPServer) {
	// Start stdio server (default mode)
	log.Printf("Project MCP Server starting in stdio mode")
	log.Printf("Set HTTP_MODE=true to run in HTTP mode instead")
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
