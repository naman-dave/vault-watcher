package vaultwatcher

import (
	"testing"
	"time"
)

// TestVaultConfig creates a test vault configuration for testing
func TestVaultConfig() *VaultConfig {
	return &VaultConfig{
		Host:  "https://test-vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token-123",
	}
}

// TestWatcher creates a test watcher instance with default configuration
func TestWatcher(t *testing.T, onChange func() error) *Watcher {
	if onChange == nil {
		onChange = func() error { return nil }
	}
	
	watcher, err := NewWatcher(TestVaultConfig(), 100*time.Millisecond, onChange)
	if err != nil {
		t.Fatalf("Failed to create test watcher: %v", err)
	}
	
	return watcher
}

// TestWatcherWithConfig creates a test watcher with custom configuration
func TestWatcherWithConfig(t *testing.T, config *VaultConfig, interval time.Duration, onChange func() error) *Watcher {
	if onChange == nil {
		onChange = func() error { return nil }
	}
	
	watcher, err := NewWatcher(config, interval, onChange)
	if err != nil {
		t.Fatalf("Failed to create test watcher with config: %v", err)
	}
	
	return watcher
}

// AssertStringEquals asserts that two strings are equal
func AssertStringEquals(t *testing.T, got, want, context string) {
	if got != want {
		t.Errorf("%s: got %q, want %q", context, got, want)
	}
}

// AssertBoolEquals asserts that two booleans are equal
func AssertBoolEquals(t *testing.T, got, want bool, context string) {
	if got != want {
		t.Errorf("%s: got %v, want %v", context, got, want)
	}
}

// AssertNoError asserts that error is nil
func AssertNoError(t *testing.T, err error, context string) {
	if err != nil {
		t.Errorf("%s: unexpected error: %v", context, err)
	}
}

// AssertError asserts that error is not nil and optionally checks the message
func AssertError(t *testing.T, err error, expectedMsg, context string) {
	if err == nil {
		t.Errorf("%s: expected error but got none", context)
		return
	}
	if expectedMsg != "" && err.Error() != expectedMsg {
		t.Errorf("%s: error message got %q, want %q", context, err.Error(), expectedMsg)
	}
}

// MockVaultData creates test vault data for hash calculations
func MockVaultData() map[string]interface{} {
	return map[string]interface{}{
		"database_url":      "postgres://localhost:5432/testdb",
		"api_key":          "test-api-key-123",
		"debug_mode":       true,
		"max_connections":  10,
		"timeout_seconds":  30.5,
		"features": []interface{}{
			"feature1",
			"feature2",
		},
		"nested_config": map[string]interface{}{
			"cache_enabled": true,
			"cache_ttl":     300,
		},
	}
}

// MockVaultDataModified creates modified test vault data
func MockVaultDataModified() map[string]interface{} {
	return map[string]interface{}{
		"database_url":      "postgres://localhost:5432/testdb",
		"api_key":          "changed-api-key-456", // Changed value
		"debug_mode":       false,                 // Changed value
		"max_connections":  20,                    // Changed value
		"timeout_seconds":  45.0,                  // Changed value
		"new_feature":      "added",               // New key
		"features": []interface{}{
			"feature1",
			"feature2",
			"feature3", // Added element
		},
		"nested_config": map[string]interface{}{
			"cache_enabled": false, // Changed value
			"cache_ttl":     600,   // Changed value
			"cache_size":    1000,  // New key
		},
	}
}