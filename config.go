package main

import (
	"fmt"
	"os"
	"time"
)

// Config holds application configuration
type Config struct {
	CerebroAPIBaseURL string
	HTTPTimeout       time.Duration
	ServerPort        string
	MCPEndpoint       string
	CerebroToken      string
	HTTPMode          bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cerebroToken := os.Getenv("CEREBRO_TOKEN")
	if cerebroToken == "" {
		return nil, fmt.Errorf("CEREBRO_TOKEN environment variable is required")
	}

	return &Config{
		CerebroAPIBaseURL: "https://cerebro.zende.sk/projects.json",
		HTTPTimeout:       30 * time.Second,
		ServerPort:        getEnvOrDefault("SERVER_PORT", ":8080"),
		MCPEndpoint:       getEnvOrDefault("MCP_ENDPOINT", "/mcp"),
		CerebroToken:      cerebroToken,
		HTTPMode:          os.Getenv("HTTP_MODE") == "true",
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
