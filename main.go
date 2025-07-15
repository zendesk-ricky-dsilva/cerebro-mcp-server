package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Constants
const (
	CerebroAPIBaseURL = "https://cerebro.zende.sk/projects.json"
	HTTPTimeout       = 30 * time.Second
	ServerPort        = ":8080"
	MCPEndpoint       = "/mcp"
)

type CerebroAPIParameters struct {
	searchKey   string
	searchValue string
	inlines     []string
	includes    []string
}

// Repository represents a repository in the API response
type Repository struct {
	ID              int     `json:"id"`
	OpenSource      bool    `json:"open_source"`
	Fork            bool    `json:"fork"`
	URL             string  `json:"url"`
	Permalink       string  `json:"permalink"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	DeprecatedOn    *string `json:"deprecated_on"`
	SyncedAt        string  `json:"synced_at"`
	StartedOn       string  `json:"started_on"`
	GithubSyncError bool    `json:"github_sync_error"`
	KubeProject     *string `json:"kube_project"`
	Archived        bool    `json:"archived"`
	Name            string  `json:"name"`
	DeletedAt       *string `json:"deleted_at"`
	Category        string  `json:"category"`
}

// Project represents a project in the API response
type Project struct {
	ID                              int      `json:"id"`
	Name                            string   `json:"name"`
	Permalink                       string   `json:"permalink"`
	Description                     string   `json:"description"`
	StartedOn                       string   `json:"started_on"`
	CreatedAt                       string   `json:"created_at"`
	UpdatedAt                       string   `json:"updated_at"`
	SlackChannel                    string   `json:"slack_channel"`
	Nickname                        string   `json:"nickname"`
	CPUUsage                        string   `json:"cpu_usage"`
	MemoryUsage                     string   `json:"memory_usage"`
	Category                        string   `json:"category"`
	DeployTarget                    string   `json:"deploy_target"`
	InScopeForSOC2                  string   `json:"in_scope_for_soc2"`
	RunsOn                          string   `json:"runs_on"`
	TFA                             string   `json:"tfa"`
	CriticalityTier                 string   `json:"criticality_tier"`
	CalculatedCriticalityTier       string   `json:"calculated_criticality_tier"`
	ReleaseState                    string   `json:"release_state"`
	LinkRepositoryURLs              []string `json:"link_repository_urls"`
	ProjectRepositoryURLs           []string `json:"project_repository_urls"`
	ProjectStakeholderOwner         string   `json:"project_stakeholder_owner_name"`
	ProjectStakeholderOncall        string   `json:"project_stakeholder_oncall_name"`
	DependentProjectDependenciesIds []int    `json:"dependent_project_dependencies_ids"`
}

// ProjectDependency represents a project dependency in the API response
type ProjectDependency struct {
	ID                 int     `json:"id"`
	DependentProjectID int     `json:"dependent_project_id"`
	ProvidingProjectID int     `json:"providing_project_id"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	Description        string  `json:"description"`
	Optional           bool    `json:"optional"`
	DeletedAt          *string `json:"deleted_at"`
}

// APIResponse represents the complete API response
type APIResponse struct {
	Pagination          map[string]interface{} `json:"pagination"`
	Projects            []Project              `json:"projects"`
	Repositories        []Repository           `json:"repositories"`
	ProjectDependencies []ProjectDependency    `json:"project_dependencies"`
}

// ProjectServer represents the MCP server
type ProjectServer struct {
	// Add any configuration fields here as needed
}

// Common helper functions

// validateProjectPermalink validates and extracts the project_permalink from request arguments
func validateProjectPermalink(arguments map[string]interface{}) (string, error) {
	projectPermalink, ok := arguments["project_permalink"].(string)
	if !ok || projectPermalink == "" {
		return "", fmt.Errorf("project_permalink parameter is required")
	}
	return projectPermalink, nil
}

// getCerebroToken gets and validates the CEREBRO_TOKEN environment variable
func getCerebroToken() (string, error) {
	cerebroToken := os.Getenv("CEREBRO_TOKEN")
	if cerebroToken == "" {
		return "", fmt.Errorf("CEREBRO_TOKEN environment variable is required")
	}
	return cerebroToken, nil
}

// buildCerebroAPIURL builds the Cerebro API URL with given parameters
func buildCerebroAPIURL(CerebroAPIParameters CerebroAPIParameters) string {
	params := url.Values{}
	params.Set(fmt.Sprintf("search[%s]", CerebroAPIParameters.searchKey), CerebroAPIParameters.searchValue)

	if len(CerebroAPIParameters.includes) > 0 {
		params.Set("includes", CerebroAPIParameters.includes[0])
	}

	if len(CerebroAPIParameters.inlines) > 0 {
		params.Set("inlines", CerebroAPIParameters.inlines[0])
	}

	return fmt.Sprintf("%s?%s", CerebroAPIBaseURL, params.Encode())
}

// makeAuthenticatedRequest makes an authenticated HTTP request to the Cerebro API
func makeAuthenticatedRequest(apiURL, token string) (*APIResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Accept", "application/json")

	// Execute request with timeout
	client := &http.Client{Timeout: HTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the JSON response
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &apiResponse, nil
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
func NewProjectServer() (*ProjectServer, error) {
	ps := &ProjectServer{
		// Initialize any fields here
	}

	return ps, nil
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

	// Validate arguments using common function
	projectPermalink, err := validateProjectPermalink(arguments)
	if err != nil {
		return nil, err
	}

	// Get Cerebro token using common function
	cerebroToken, err := getCerebroToken()
	if err != nil {
		return nil, err
	}

	// Build Cerebro API URL with project repositories included
	ProjectDetailsQuery := CerebroAPIParameters{
		searchKey:   "permalink",
		searchValue: projectPermalink,
		inlines:     []string{"project_repository_urls", "project_stakeholder_owner_name", "project_stakeholder_oncall_name", "link_deployment_url", "link_deployment_urls"},
		includes:    []string{"project_repositories"},
	}
	apiURL := buildCerebroAPIURL(ProjectDetailsQuery)

	// Make authenticated request using common function
	apiResponse, err := makeAuthenticatedRequest(apiURL, cerebroToken)
	if err != nil {
		return nil, err
	}

	// Filter repositories where kube_project matches the project_permalink
	var matchingRepos []Repository
	for _, repo := range apiResponse.Repositories {
		if repo.KubeProject != nil && *repo.KubeProject == projectPermalink {
			matchingRepos = append(matchingRepos, repo)
		}
	}

	// Format the response
	result := fmt.Sprintf("# Project Details for: %s\n\n", projectPermalink)

	if len(apiResponse.Projects) > 0 {
		project := apiResponse.Projects[0]
		result += fmt.Sprintf("**Project Name:** %s\n", project.Name)
		result += fmt.Sprintf("**Description:** %s\n", project.Description)
		result += fmt.Sprintf("**Category:** %s\n", project.Category)
		result += fmt.Sprintf("**Calculated Criticality Tier:** %s\n", project.CalculatedCriticalityTier)
		result += fmt.Sprintf("**Release State:** %s\n", project.ReleaseState)
		result += fmt.Sprintf("**Owner:** %s\n", project.ProjectStakeholderOwner)
		result += fmt.Sprintf("**Slack Channel:** %s\n", project.SlackChannel)

		// Add project repository URLs if available
		if len(project.ProjectRepositoryURLs) > 0 {
			result += fmt.Sprintf("\n**Project Repository URLs (%d):**\n", len(project.ProjectRepositoryURLs))
			for i, repoURL := range project.ProjectRepositoryURLs {
				result += fmt.Sprintf("%d. %s\n", i+1, repoURL)
			}
		}

		result += "\n"
	}

	if len(matchingRepos) == 0 {
		result += "No repositories found with matching kube_project.\n"
	} else {
		result += fmt.Sprintf("%d repositories linked to kube_project '%s'\n\n", len(matchingRepos), projectPermalink)

		for i, repo := range matchingRepos {
			result += fmt.Sprintf("### %d. %s\n", i+1, repo.Name)
			result += fmt.Sprintf("- **Permalink:** %s\n", repo.Permalink)
			result += fmt.Sprintf("- **URL:** %s\n", repo.URL)
			result += fmt.Sprintf("- **Archived:** %t\n", repo.Archived)
			if repo.DeprecatedOn != nil {
				result += fmt.Sprintf("- **Deprecated On:** %s\n", *repo.DeprecatedOn)
			}
			result += fmt.Sprintf("- **Last Updated:** %s\n\n", repo.UpdatedAt)
		}
	}

	return createMCPResult(result), nil
}

func (ps *ProjectServer) handleGetProjectDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()

	// Validate arguments using common function
	projectPermalink, err := validateProjectPermalink(arguments)
	if err != nil {
		return nil, err
	}

	// Get Cerebro token using common function
	cerebroToken, err := getCerebroToken()
	if err != nil {
		return nil, err
	}

	// First, get the project to find its ID and dependencies

	// Build Cerebro API URL with project dependencies included
	ProjectDependenciesQuery := CerebroAPIParameters{
		searchKey:   "permalink",
		searchValue: projectPermalink,
		inlines:     []string{"project_repository_urls", "project_stakeholder_owner_name", "project_stakeholder_oncall_name", "link_deployment_url", "link_deployment_urls"},
		includes:    []string{"dependent_project_dependencies"},
	}
	apiURL := buildCerebroAPIURL(ProjectDependenciesQuery)

	projectDependenciesResponse, err := makeAuthenticatedRequest(apiURL, cerebroToken)
	if err != nil {
		return nil, err
	}

	// Check if project exists
	if len(projectDependenciesResponse.Projects) == 0 {
		return nil, fmt.Errorf("project not found: %s", projectPermalink)
	}

	project := projectDependenciesResponse.Projects[0]
	projectID := project.ID

	// If no dependencies are found, return early
	if len(projectDependenciesResponse.ProjectDependencies) == 0 {
		return createMCPResult(fmt.Sprintf("# No dependencies found for project: %s\n\n", project.Name)), nil
	}

	// Filter dependencies where this project is the dependent
	var relevantDependencies []ProjectDependency
	for _, dep := range projectDependenciesResponse.ProjectDependencies {
		if dep.DependentProjectID == projectID {
			relevantDependencies = append(relevantDependencies, dep)
		}
	}

	// Get details for each providing project
	result := fmt.Sprintf("# Dependencies for Project: %s\n\n", project.Name)
	result += fmt.Sprintf("**Project ID:** %d\n", project.ID)
	result += fmt.Sprintf("**Permalink:** %s\n", project.Permalink)
	result += fmt.Sprintf("**Description:** %s\n\n", project.Description)
	result += fmt.Sprintf("## Dependencies (%d)\n\n", len(relevantDependencies))

	// Create a channel to collect results
	type dependencyResult struct {
		index            int
		dep              ProjectDependency
		providingProject *Project
		err              error
	}

	results := make(chan dependencyResult, len(relevantDependencies))

	// Launch goroutines for each dependency
	for i, dep := range relevantDependencies {
		go func(index int, dependency ProjectDependency) {
			// Get providing project details
			ProvidingProjectDependencyQuery := CerebroAPIParameters{
				searchKey:   "id",
				searchValue: fmt.Sprintf("%d", dependency.ProvidingProjectID),
			}
			providingURL := buildCerebroAPIURL(ProvidingProjectDependencyQuery)

			providingResponse, err := makeAuthenticatedRequest(providingURL, cerebroToken)

			if err != nil {
				results <- dependencyResult{index: index, dep: dependency, err: err}
				return
			}

			if len(providingResponse.Projects) == 0 {
				results <- dependencyResult{index: index, dep: dependency, providingProject: nil, err: nil}
				return
			}

			results <- dependencyResult{index: index, dep: dependency, providingProject: &providingResponse.Projects[0], err: nil}
		}(i, dep)
	}

	// Collect all results
	dependencyResults := make([]dependencyResult, len(relevantDependencies))
	for i := 0; i < len(relevantDependencies); i++ {
		res := <-results
		dependencyResults[res.index] = res
	}

	// Process results in order
	for i, res := range dependencyResults {
		if res.err != nil {
			result += fmt.Sprintf("### %d. Error fetching project ID %d\n", i+1, res.dep.ProvidingProjectID)
			result += fmt.Sprintf("- **Error:** %v\n\n", res.err)
			continue
		}

		if res.providingProject == nil {
			result += fmt.Sprintf("### %d. Project ID %d (Not Found)\n", i+1, res.dep.ProvidingProjectID)
			result += fmt.Sprintf("- **Dependency ID:** %d\n", res.dep.ID)
			result += fmt.Sprintf("- **Optional:** %t\n\n", res.dep.Optional)
			continue
		}

		providingProject := res.providingProject
		result += fmt.Sprintf("### %d. %s\n", i+1, providingProject.Name)
		result += fmt.Sprintf("- **Dependency ID:** %d\n", res.dep.ID)
		result += fmt.Sprintf("- **Providing Project ID:** %d\n", res.dep.ProvidingProjectID)
		result += fmt.Sprintf("- **Permalink:** %s\n", providingProject.Permalink)
		result += fmt.Sprintf("- **Description:** %s\n", providingProject.Description)
		result += fmt.Sprintf("- **Category:** %s\n", providingProject.Category)
		result += fmt.Sprintf("- **Criticality Tier:** %s\n", providingProject.CalculatedCriticalityTier)
		result += fmt.Sprintf("- **Release State:** %s\n", providingProject.ReleaseState)
		result += fmt.Sprintf("- **Owner Team:** %s\n", providingProject.ProjectStakeholderOwner)
		result += fmt.Sprintf("- **Slack Channel:** %s\n", providingProject.SlackChannel)
		result += fmt.Sprintf("- **Optional Dependency:** %t\n", res.dep.Optional)
		if res.dep.Description != "" {
			result += fmt.Sprintf("- **Dependency Description:** %s\n", res.dep.Description)
		}
	}

	return createMCPResult(result), nil
}

// HTTPRequest represents a simplified HTTP request for tool calls
type HTTPRequest struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// HTTPResponse represents the HTTP response
type HTTPResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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
	if httpReq.Tool != "project_get_details" && httpReq.Tool != "project_get_dependencies" {
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
	case "project_get_details":
		result, err = ps.handleGetProjectDetails(context.Background(), mcpRequest)
	case "project_get_dependencies":
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
	// Create Project server instance
	projectServer, err := NewProjectServer()
	if err != nil {
		log.Fatalf("Failed to create Project server: %v", err)
	}

	// Setup MCP server with tools and resources
	mcpServer := projectServer.SetupMCPServer()

	// Check if HTTP_MODE environment variable is set
	if os.Getenv("HTTP_MODE") == "true" {
		// Start HTTP server
		http.Handle(MCPEndpoint, projectServer)
		log.Printf("Project MCP Server starting HTTP mode on %s", ServerPort)
		log.Printf("Send POST requests to http://localhost%s%s", ServerPort, MCPEndpoint)
		log.Printf("Available tools: project_get_details, project_get_dependencies")
		log.Printf("Example request body: {\"tool\": \"project_get_details\", \"arguments\": {\"project_permalink\": \"your-project\"}}")
		log.Printf("Example dependencies request: {\"tool\": \"project_get_dependencies\", \"arguments\": {\"project_permalink\": \"your-project\"}}")
		if err := http.ListenAndServe(ServerPort, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	} else {
		// Start stdio server (default mode)
		log.Printf("Project MCP Server starting in stdio mode")
		log.Printf("Set HTTP_MODE=true to run in HTTP mode instead")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
