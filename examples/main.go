package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/naman-dave/vault-watcher"
)

func main() {
	// Example 1: Basic usage with environment variables
	basicExample()

	// Example 2: Custom configuration
	customConfigExample()

	// Example 3: Advanced usage with custom callback
	advancedExample()
}

func basicExample() {
	fmt.Println("=== Basic Example ===")

	// Load configuration from environment variables
	config, err := vaultwatcher.LoadVaultConfigFromEnv()
	if err != nil {
		log.Printf("Failed to load config from environment: %v", err)
		log.Println("Please set VAULT_HOST, VAULT_PATH, and VAULT_TOKEN environment variables")
		return
	}

	// Simple onChange callback
	onChange := func() error {
		fmt.Println("🔄 Vault configuration changed!")
		// Add your custom logic here
		return nil
	}

	// Create watcher (checks every 30 seconds)
	watcher, err := vaultwatcher.NewWatcher(config, 30*time.Second, onChange)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	// Start monitoring
	if err := watcher.Start(); err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	fmt.Printf("Started monitoring %s every 30 seconds\n", config.Path)
	fmt.Printf("Initial hash: %s\n", watcher.GetCurrentHash())

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Stopping watcher...")
}

func customConfigExample() {
	fmt.Println("\n=== Custom Configuration Example ===")

	// Custom vault configuration
	config := &vaultwatcher.VaultConfig{
		Host:  "https://vault.company.com:8200",
		Path:  "kv/data/production/myapp",
		Token: "hvs.your-token-here",
	}

	onChange := func() error {
		fmt.Println("🚀 Production config changed!")
		// Example: restart application, reload config, etc.
		return nil
	}

	// Create watcher with custom interval (5 minutes)
	watcher, err := vaultwatcher.NewWatcher(config, 5*time.Minute, onChange)
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}

	fmt.Printf("Would monitor %s every 5 minutes\n", config.Path)
	fmt.Println("Note: This example uses dummy credentials")

	// Don't actually start for this example
	defer watcher.Stop()
}

func advancedExample() {
	fmt.Println("\n=== Advanced Example ===")

	config, err := vaultwatcher.LoadVaultConfigFromEnv()
	if err != nil {
		log.Printf("Skipping advanced example: %v", err)
		return
	}

	// Advanced onChange callback with error handling
	changeCount := 0
	onChange := func() error {
		changeCount++
		fmt.Printf("📈 Change #%d detected at %s\n", changeCount, time.Now().Format(time.RFC3339))

		// Example: Validate new configuration before applying
		if err := validateNewConfig(); err != nil {
			fmt.Printf("❌ New configuration is invalid: %v\n", err)
			return err
		}

		// Example: Apply new configuration
		if err := applyNewConfig(); err != nil {
			fmt.Printf("❌ Failed to apply new configuration: %v\n", err)
			return err
		}

		fmt.Println("✅ Configuration successfully updated")
		return nil
	}

	// Create watcher with short interval for demo
	watcher, err := vaultwatcher.NewWatcher(config, 10*time.Second, onChange)
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}

	if err := watcher.Start(); err != nil {
		log.Printf("Failed to start watcher: %v", err)
		return
	}
	defer watcher.Stop()

	fmt.Printf("Started advanced monitoring of %s\n", config.Path)
	fmt.Printf("Watcher status: %v\n", watcher.IsStarted())

	// Monitor for a short time for demo
	time.Sleep(30 * time.Second)

	fmt.Printf("Final change count: %d\n", changeCount)
	fmt.Printf("Final hash: %s\n", watcher.GetCurrentHash())
}

// Example validation function
func validateNewConfig() error {
	// Add your validation logic here
	// For example: check required fields, validate URLs, test connections, etc.
	fmt.Println("🔍 Validating new configuration...")
	time.Sleep(100 * time.Millisecond) // Simulate validation time
	return nil
}

// Example configuration application function
func applyNewConfig() error {
	// Add your configuration application logic here
	// For example: restart services, reload config files, update caches, etc.
	fmt.Println("⚙️  Applying new configuration...")
	time.Sleep(200 * time.Millisecond) // Simulate application time
	return nil
}

// Example showing how to handle different scenarios
func demonstrateErrorHandling() {
	fmt.Println("\n=== Error Handling Example ===")

	config := &vaultwatcher.VaultConfig{
		Host:  "https://invalid-vault.example.com",
		Path:  "secret/test",
		Token: "invalid-token",
	}

	onChange := func() error {
		return fmt.Errorf("simulated callback error")
	}

	// This will fail due to invalid configuration
	watcher, err := vaultwatcher.NewWatcher(config, time.Second, onChange)
	if err != nil {
		fmt.Printf("Expected error creating watcher: %v\n", err)
		return
	}
	defer watcher.Stop()

	// This would fail when trying to connect to Vault
	if err := watcher.Start(); err != nil {
		fmt.Printf("Expected error starting watcher: %v\n", err)
	}
}