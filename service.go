package main

import (
	"context"
	"fmt"
	"sync"
)

// ProjectService handles project-related business logic
type ProjectService struct {
	client    *CerebroClient
	validator *Validator
}

// NewProjectService creates a new ProjectService
func NewProjectService(client *CerebroClient, validator *Validator) *ProjectService {
	return &ProjectService{
		client:    client,
		validator: validator,
	}
}

// GetProjectDetails retrieves detailed information about a project
func (s *ProjectService) GetProjectDetails(ctx context.Context, permalink string) (*ProjectDetailsResult, error) {
	if err := s.validator.ValidateProjectPermalink(permalink); err != nil {
		return nil, err
	}

	params := CerebroAPIParameters{
		searchKey:   "permalink",
		searchValue: permalink,
		inlines:     "project_repository_urls,project_stakeholder_owner_name,project_stakeholder_oncall_name,link_deployment_url,link_deployment_urls",
	}

	apiURL := s.client.buildURL(params)
	response, err := s.client.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, err
	}

	if len(response.Projects) == 0 {
		return nil, &ProjectNotFoundError{Permalink: permalink}
	}

	project := response.Projects[0]

	return &ProjectDetailsResult{
		Project:       project,
		FormattedText: s.formatProjectDetails(project, permalink),
	}, nil
}

// GetProjectDependencies retrieves dependency information for a project
func (s *ProjectService) GetProjectDependencies(ctx context.Context, permalink string) (*ProjectDependenciesResult, error) {
	if err := s.validator.ValidateProjectPermalink(permalink); err != nil {
		return nil, err
	}

	params := CerebroAPIParameters{
		searchKey:   "permalink",
		searchValue: permalink,
		inlines:     "project_repository_urls,project_stakeholder_owner_name,project_stakeholder_oncall_name,link_deployment_url,link_deployment_urls",
		includes:    "dependent_project_dependencies",
	}

	apiURL := s.client.buildURL(params)
	response, err := s.client.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, err
	}

	if len(response.Projects) == 0 {
		return nil, &ProjectNotFoundError{Permalink: permalink}
	}

	project := response.Projects[0]
	if len(response.ProjectDependencies) == 0 {
		return &ProjectDependenciesResult{
			Project:             project,
			ProjectDependencies: []ProjectDependency{},
			FormattedText:       fmt.Sprintf("# No dependencies found for project: %s\n\n", project.Name),
		}, nil
	}

	relevantDependencies := s.filterDependencies(response.ProjectDependencies, project.ID)
	dependenciesWithDetails := s.fetchDependenciesAsync(ctx, relevantDependencies)

	return &ProjectDependenciesResult{
		Project:             project,
		ProjectDependencies: relevantDependencies,
		FormattedText:       s.formatDependencies(project, dependenciesWithDetails),
	}, nil
}

// filterDependencies filters dependencies where the project is the dependent
func (s *ProjectService) filterDependencies(deps []ProjectDependency, projectID int) []ProjectDependency {
	var relevant []ProjectDependency
	for _, dep := range deps {
		if dep.DependentProjectID == projectID {
			relevant = append(relevant, dep)
		}
	}
	return relevant
}

// dependencyResult represents the result of fetching a single dependency
type dependencyResult struct {
	index            int
	dep              ProjectDependency
	providingProject *Project
	err              error
}

// fetchDependenciesAsync fetches dependency details asynchronously
func (s *ProjectService) fetchDependenciesAsync(ctx context.Context, dependencies []ProjectDependency) []dependencyResult {
	results := make([]dependencyResult, len(dependencies))
	var wg sync.WaitGroup

	for i, dep := range dependencies {
		wg.Add(1)
		go func(index int, dependency ProjectDependency) {
			defer wg.Done()

			params := CerebroAPIParameters{
				searchKey:   "id",
				searchValue: fmt.Sprintf("%d", dependency.ProvidingProjectID),
			}

			apiURL := s.client.buildURL(params)
			response, err := s.client.makeRequest(ctx, apiURL)

			if err != nil {
				results[index] = dependencyResult{index: index, dep: dependency, err: err}
				return
			}

			if len(response.Projects) == 0 {
				results[index] = dependencyResult{index: index, dep: dependency, providingProject: nil, err: nil}
				return
			}

			results[index] = dependencyResult{index: index, dep: dependency, providingProject: &response.Projects[0], err: nil}
		}(i, dep)
	}

	wg.Wait()
	return results
}

// formatProjectDetails formats project details for display
func (s *ProjectService) formatProjectDetails(project Project, permalink string) string {
	result := fmt.Sprintf("# Project Details for: %s\n\n", permalink)

	result += fmt.Sprintf("**Project Name:** %s\n", project.Name)
	result += fmt.Sprintf("**Description:** %s\n", project.Description)
	result += fmt.Sprintf("**Category:** %s\n", project.Category)
	result += fmt.Sprintf("**Calculated Criticality Tier:** %s\n", project.CalculatedCriticalityTier)
	result += fmt.Sprintf("**Release State:** %s\n", project.ReleaseState)
	result += fmt.Sprintf("**Owner:** %s\n", project.ProjectStakeholderOwner)
	result += fmt.Sprintf("**Slack Channel:** %s\n", project.SlackChannel)

	if len(project.ProjectRepositoryURLs) == 0 {
		result += "No project repository URLs found.\n"
	} else {
		result += fmt.Sprintf("\n**Project Repository URLs (%d):**\n", len(project.ProjectRepositoryURLs))
		for i, repoURL := range project.ProjectRepositoryURLs {
			result += fmt.Sprintf("%d. %s\n", i+1, repoURL)
		}
	}

	return result
}

// formatDependencies formats dependencies for display
func (s *ProjectService) formatDependencies(project Project, dependencyResults []dependencyResult) string {
	result := fmt.Sprintf("# Dependencies for Project: %s\n\n", project.Name)
	result += fmt.Sprintf("**Project ID:** %d\n", project.ID)
	result += fmt.Sprintf("**Permalink:** %s\n", project.Permalink)
	result += fmt.Sprintf("**Description:** %s\n\n", project.Description)
	result += fmt.Sprintf("## Dependencies (%d)\n\n", len(dependencyResults))

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
		result += "\n"
	}

	return result
}
