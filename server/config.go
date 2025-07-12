package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type AgentConfig struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

type Config struct {
	Agents []AgentConfig `toml:"agents"`
}

// LoadConfig loads configuration from TOML file
func LoadConfig(configPath string) (*Config, error) {
	// Try to find config file
	if configPath == "" {
		configPath = "./config.toml"
	}

	// If not found in current directory, try relative to executable
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		execDir, err := os.Executable()
		if err == nil {
			serverDir := filepath.Dir(execDir)
			configPath = filepath.Join(serverDir, "..", "config.toml")
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Replace environment variables in URLs
	for i := range config.Agents {
		config.Agents[i].URL = replaceEnvVars(config.Agents[i].URL)
	}

	return &config, nil
}

// replaceEnvVars replaces ${VAR_NAME} patterns with environment variable values
func replaceEnvVars(s string) string {
	// Simple replacement for ${VAR_NAME} pattern
	if strings.Contains(s, "${") && strings.Contains(s, "}") {
		start := strings.Index(s, "${")
		end := strings.Index(s, "}")
		if start != -1 && end != -1 && end > start {
			varName := s[start+2 : end]
			envValue := os.Getenv(varName)
			return s[:start] + envValue + s[end+1:]
		}
	}
	return s
}

// GetAgentURLs returns a slice of agent URLs from the configuration
func (c *Config) GetAgentURLs() []string {
	urls := make([]string, len(c.Agents))
	for i, agent := range c.Agents {
		urls[i] = agent.URL
	}
	return urls
}
