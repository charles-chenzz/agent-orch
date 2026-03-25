// Package config handles application configuration.
package config

import (
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Terminal TerminalConfig `mapstructure:"terminal"`
	Proxy    ProxyConfig    `mapstructure:"proxy"`
	Profiles []Profile      `mapstructure:"profiles"`
	GitHub   GitHubConfig   `mapstructure:"github"`
}

// AppConfig holds application settings.
type AppConfig struct {
	Theme       string `mapstructure:"theme"`
	CheckUpdate bool   `mapstructure:"check_updates"`
}

// TerminalConfig holds terminal settings.
type TerminalConfig struct {
	Shell      string `mapstructure:"shell"`
	FontFamily string `mapstructure:"font_family"`
	FontSize   int    `mapstructure:"font_size"`
}

// ProxyConfig holds proxy settings.
type ProxyConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

// Profile holds API provider profile.
type Profile struct {
	Name     string `mapstructure:"name"`
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Active   bool   `mapstructure:"active"`
}

// GitHubConfig holds GitHub integration settings.
type GitHubConfig struct {
	Token        string `mapstructure:"token"`
	AutoDetectPR bool   `mapstructure:"auto_detect_pr"`
}

// Manager handles configuration operations.
type Manager struct {
	cfg  *Config
	path string
}

// Load loads configuration from file.
func Load() (*Manager, error) {
	configPath := getConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Default(), nil
	}

	// TODO: Use viper to load config
	return Default(), nil
}

// Default returns default configuration.
func Default() *Manager {
	cfg := &Config{
		App: AppConfig{
			Theme:       "dark",
			CheckUpdate: true,
		},
		Terminal: TerminalConfig{
			Shell:      "/bin/zsh",
			FontFamily: "JetBrains Mono",
			FontSize:   14,
		},
		Proxy: ProxyConfig{
			Enabled: true,
			Port:    8080,
		},
		Profiles: []Profile{},
		GitHub: GitHubConfig{
			AutoDetectPR: true,
		},
	}

	return &Manager{cfg: cfg, path: getConfigPath()}
}

// Current returns current configuration.
func (m *Manager) Current() *Config {
	return m.cfg
}

// Save saves configuration to file.
func (m *Manager) Save(cfg *Config) error {
	m.cfg = cfg
	// TODO: Implement save with viper
	return nil
}

// DatabasePath returns the database file path.
func (m *Manager) DatabasePath() string {
	return filepath.Join(m.AppDir(), "data.db")
}

// AppDir returns the application directory (~/.agent-orch/)
func (m *Manager) AppDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agent-orch")
}

// WorktreeBaseDir returns the base directory for all worktrees
// Structure: ~/.agent-orch/worktrees/<project-name>/<worktree-name>/
func (m *Manager) WorktreeBaseDir() string {
	return filepath.Join(m.AppDir(), "worktrees")
}

// ProjectWorktreeDir returns the worktree directory for a specific project
func (m *Manager) ProjectWorktreeDir(projectName string) string {
	return filepath.Join(m.WorktreeBaseDir(), projectName)
}

// EnsureAppDir ensures the application directory structure exists
func (m *Manager) EnsureAppDir() error {
	dirs := []string{
		m.AppDir(),
		m.WorktreeBaseDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agent-orch", "config.toml")
}
