package config

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestConfig creates a temporary config file for testing
func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "a-la-carte-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a temporary config file
	configPath := filepath.Join(tempDir, "a-la-carte.yml")
	configContent := `
ui:
  theme: dark
  detailHeight: 15
  listHeight: 20

software:
  manifestPath: test-manifest.yml
  preloadKeys:
    - git
    - vim

system:
  debugMode: true
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return configPath, cleanup
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default values
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got %s", cfg.UI.Theme)
	}

	if cfg.UI.DetailHeight != 10 {
		t.Errorf("expected default detail height 10, got %d", cfg.UI.DetailHeight)
	}

	if cfg.UI.ListHeight != 10 {
		t.Errorf("expected default list height 10, got %d", cfg.UI.ListHeight)
	}

	if cfg.Software.ManifestPath != "software.yml" {
		t.Errorf("expected default manifest path 'software.yml', got %s", cfg.Software.ManifestPath)
	}

	if cfg.System.DebugMode != false {
		t.Errorf("expected default debug mode 'false', got %v", cfg.System.DebugMode)
	}

	if len(cfg.Software.PreloadKeys) != 0 {
		t.Errorf("expected empty preload keys, got %v", cfg.Software.PreloadKeys)
	}
}

func TestLoad(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify loaded values
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %s", cfg.UI.Theme)
	}

	if cfg.UI.DetailHeight != 15 {
		t.Errorf("expected detail height 15, got %d", cfg.UI.DetailHeight)
	}

	if cfg.UI.ListHeight != 20 {
		t.Errorf("expected list height 20, got %d", cfg.UI.ListHeight)
	}

	if cfg.Software.ManifestPath != "test-manifest.yml" {
		t.Errorf("expected manifest path 'test-manifest.yml', got %s", cfg.Software.ManifestPath)
	}

	if cfg.System.DebugMode != true {
		t.Errorf("expected debug mode 'true', got %v", cfg.System.DebugMode)
	}

	if len(cfg.Software.PreloadKeys) != 2 || cfg.Software.PreloadKeys[0] != "git" || cfg.Software.PreloadKeys[1] != "vim" {
		t.Errorf("expected preload keys ['git', 'vim'], got %v", cfg.Software.PreloadKeys)
	}

	if cfg.ConfigPath != configPath {
		t.Errorf("expected config path %s, got %s", configPath, cfg.ConfigPath)
	}
}

func TestLoadError(t *testing.T) {
	// Test with non-existent file
	_, err := Load("non-existent-file.yml")
	if err == nil {
		t.Fatal("expected error loading non-existent file, got nil")
	}

	// Test with invalid YAML
	tempFile, err := os.CreateTemp("", "invalid-config-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("invalid: yaml: :")
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	_, err = Load(tempFile.Name())
	if err == nil {
		t.Fatal("expected error loading invalid YAML, got nil")
	}
}

func TestValidate(t *testing.T) {
	// Test valid config
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no validation error for default config, got %v", err)
	}

	// Test invalid theme
	cfg.UI.Theme = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid theme, got nil")
	}

	// Reset and test invalid detail height
	cfg = DefaultConfig()
	cfg.UI.DetailHeight = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid detail height, got nil")
	}

	// Reset and test invalid list height
	cfg = DefaultConfig()
	cfg.UI.ListHeight = -1
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid list height, got nil")
	}

	// Reset and test empty manifest path
	cfg = DefaultConfig()
	cfg.Software.ManifestPath = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty manifest path, got nil")
	}
}

func TestSave(t *testing.T) {
	// Create a temporary directory for saving
	tempDir, err := os.MkdirTemp("", "a-la-carte-save-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	savePath := filepath.Join(tempDir, "saved-config.yml")

	// Create a config to save
	cfg := DefaultConfig()
	cfg.UI.Theme = "light"
	cfg.Software.PreloadKeys = []string{"test1", "test2"}

	// Save the config
	err = cfg.Save(savePath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load the saved config
	loadedCfg, err := Load(savePath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	// Verify the saved values
	if loadedCfg.UI.Theme != "light" {
		t.Errorf("expected theme 'light', got %s", loadedCfg.UI.Theme)
	}

	if len(loadedCfg.Software.PreloadKeys) != 2 ||
		loadedCfg.Software.PreloadKeys[0] != "test1" ||
		loadedCfg.Software.PreloadKeys[1] != "test2" {
		t.Errorf("expected preload keys ['test1', 'test2'], got %v", loadedCfg.Software.PreloadKeys)
	}
}
