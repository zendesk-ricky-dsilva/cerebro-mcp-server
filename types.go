package main

// CerebroAPIParameters represents the parameters for Cerebro API requests
type CerebroAPIParameters struct {
	searchKey   string
	searchValue string
	inlines     string
	includes    string
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
	PrimaryDeploymentUrl            string   `json:"link_deployment_url"`
	AdditionalDeploymentUrls        []string `json:"link_deployment_urls"`
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
	ProjectDependencies []ProjectDependency    `json:"project_dependencies"`
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

// ProjectDetailsResult represents the result of project details query
type ProjectDetailsResult struct {
	Project       Project
	FormattedText string
}

// ProjectDependenciesResult represents the result of project dependencies query
type ProjectDependenciesResult struct {
	Project             Project
	ProjectDependencies []ProjectDependency
	FormattedText       string
}
