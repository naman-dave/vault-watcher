// +build integration

package vaultwatcher

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// These tests require a running Vault instance and proper environment variables.
// To run integration tests:
//   go test -tags=integration -v
//
// Required environment variables:
//   VAULT_HOST - Vault server address (e.g., http://localhost:8200)
//   VAULT_PATH - Path to test secret (e.g., secret/test)
//   VAULT_TOKEN - Valid Vault token
//
// Example setup with Vault dev server:
//   vault server -dev
//   export VAULT_ADDR='http://127.0.0.1:8200'
//   export VAULT_TOKEN="dev-only-token"
//   vault kv put secret/test key1=value1 key2=value2

func TestIntegration_WatcherWithRealVault(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load config from environment
	config, err := LoadVaultConfigFromEnv()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	t.Logf("Testing with Vault at %s, path %s", config.Host, config.Path)

	changeCount := 0
	onChange := func() error {
		changeCount++
		t.Logf("Change detected! Count: %d", changeCount)
		return nil
	}

	// Create watcher with short interval for testing
	watcher, err := NewWatcher(config, 2*time.Second, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	initialHash := watcher.GetCurrentHash()
	if initialHash == "" {
		t.Error("Initial hash should not be empty")
	}

	t.Logf("Initial hash: %s", initialHash)
	t.Logf("Watcher started. To test change detection, modify the secret at %s", config.Path)
	t.Logf("Example: vault kv patch %s key1=modified_value", config.Path)

	// Wait for a few monitoring cycles
	time.Sleep(10 * time.Second)

	// Check if watcher is still running
	if !watcher.IsStarted() {
		t.Error("Watcher should still be running")
	}

	currentHash := watcher.GetCurrentHash()
	t.Logf("Final hash: %s", currentHash)
	
	if changeCount > 0 {
		t.Logf("SUCCESS: Detected %d changes during test", changeCount)
		if currentHash == initialHash {
			t.Error("Hash should have changed if changes were detected")
		}
	} else {
		t.Logf("No changes detected (this is expected if secret wasn't modified)")
		if currentHash != initialHash {
			t.Error("Hash should not have changed if no changes were detected")
		}
	}
}

func TestIntegration_FetchVaultData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config, err := LoadVaultConfigFromEnv()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Create watcher to test vault data fetching
	watcher, err := NewWatcher(config, time.Minute, func() error { return nil })
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Test fetching vault data
	data, err := watcher.fetchVaultData()
	if err != nil {
		t.Fatalf("Failed to fetch vault data: %v", err)
	}

	if data == nil {
		t.Fatal("Vault data should not be nil")
	}

	t.Logf("Fetched %d keys from Vault:", len(data))
	for key, value := range data {
		t.Logf("  %s: %v", key, value)
	}

	// Test hash calculation with real data
	hash, err := CalculateHash(data)
	if err != nil {
		t.Fatalf("Failed to calculate hash: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	t.Logf("Calculated hash: %s", hash)

	// Test hash consistency
	hash2, err := CalculateHash(data)
	if err != nil {
		t.Fatalf("Failed to calculate second hash: %v", err)
	}

	if hash != hash2 {
		t.Error("Hash should be consistent")
	}
}

func TestIntegration_EnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	requiredVars := []string{"VAULT_HOST", "VAULT_PATH", "VAULT_TOKEN"}
	
	for _, varName := range requiredVars {
		value := os.Getenv(varName)
		if value == "" {
			t.Errorf("Required environment variable %s is not set", varName)
		} else {
			t.Logf("%s is set (length: %d)", varName, len(value))
		}
	}
}

// Example test that demonstrates how to test with manually set data
func TestIntegration_ManualVaultTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test shows how you might set up specific test data in Vault
	// and then test the watcher behavior
	
	config, err := LoadVaultConfigFromEnv()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// You could use vault client here to set up specific test data
	// For now, we'll just demonstrate the testing pattern

	changeDetected := false
	onChange := func() error {
		changeDetected = true
		t.Log("Change callback triggered!")
		return nil
	}

	watcher, err := NewWatcher(config, time.Second, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// In a real scenario, you might:
	// 1. Record initial state
	// 2. Modify the vault data programmatically  
	// 3. Wait for change detection
	// 4. Verify the change was detected

	initialHash := watcher.GetCurrentHash()
	t.Logf("Initial hash: %s", initialHash)

	// Wait a short time to ensure watcher is running
	time.Sleep(3 * time.Second)

	// Log current state
	t.Logf("Change detected: %v", changeDetected)
	t.Logf("Watcher running: %v", watcher.IsStarted())
	t.Logf("Current hash: %s", watcher.GetCurrentHash())
}

// Benchmark test with real Vault
func BenchmarkIntegration_HashCalculation(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration benchmark in short mode")
	}

	config, err := LoadVaultConfigFromEnv()
	if err != nil {
		b.Skipf("Skipping integration benchmark: %v", err)
	}

	watcher, err := NewWatcher(config, time.Minute, func() error { return nil })
	if err != nil {
		b.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Fetch data once
	data, err := watcher.fetchVaultData()
	if err != nil {
		b.Fatalf("Failed to fetch vault data: %v", err)
	}

	b.ResetTimer()
	
	// Benchmark hash calculation
	for i := 0; i < b.N; i++ {
		_, err := CalculateHash(data)
		if err != nil {
			b.Fatalf("Hash calculation failed: %v", err)
		}
	}
}

// Helper function to print test instructions
func init() {
	if os.Getenv("PRINT_INTEGRATION_HELP") == "true" {
		fmt.Println(`
Integration Test Setup Instructions:

1. Start Vault dev server:
   vault server -dev

2. Set environment variables:
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_HOST='http://127.0.0.1:8200'
   export VAULT_PATH='secret/test'
   export VAULT_TOKEN='<your-dev-token>'

3. Create test secret:
   vault kv put secret/test key1=value1 key2=value2

4. Run integration tests:
   go test -tags=integration -v

5. To test change detection, modify the secret while tests are running:
   vault kv patch secret/test key1=modified_value
`)
	}
}