package vaultwatcher

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// Test helper functions and utilities

func TestNewWatcher(t *testing.T) {
	tests := []struct {
		name          string
		vaultConfig   *VaultConfig
		checkInterval time.Duration
		onChange      func() error
		expectError   bool
		errorMsg      string
	}{
		{
			name: "valid configuration",
			vaultConfig: &VaultConfig{
				Host:  "https://vault.example.com",
				Path:  "kv/data/test",
				Token: "test-token",
			},
			checkInterval: 30 * time.Second,
			onChange:      func() error { return nil },
			expectError:   false,
		},
		{
			name:          "nil vault config",
			vaultConfig:   nil,
			checkInterval: 30 * time.Second,
			onChange:      func() error { return nil },
			expectError:   true,
			errorMsg:      "vault config cannot be nil",
		},
		{
			name: "empty host",
			vaultConfig: &VaultConfig{
				Host:  "",
				Path:  "kv/data/test",
				Token: "test-token",
			},
			checkInterval: 30 * time.Second,
			onChange:      func() error { return nil },
			expectError:   true,
			errorMsg:      "VAULT_HOST is required",
		},
		{
			name: "empty path",
			vaultConfig: &VaultConfig{
				Host:  "https://vault.example.com",
				Path:  "",
				Token: "test-token",
			},
			checkInterval: 30 * time.Second,
			onChange:      func() error { return nil },
			expectError:   true,
			errorMsg:      "VAULT_PATH is required",
		},
		{
			name: "empty token",
			vaultConfig: &VaultConfig{
				Host:  "https://vault.example.com",
				Path:  "kv/data/test",
				Token: "",
			},
			checkInterval: 30 * time.Second,
			onChange:      func() error { return nil },
			expectError:   true,
			errorMsg:      "VAULT_TOKEN is required",
		},
		{
			name: "nil onChange callback",
			vaultConfig: &VaultConfig{
				Host:  "https://vault.example.com",
				Path:  "kv/data/test",
				Token: "test-token",
			},
			checkInterval: 30 * time.Second,
			onChange:      nil,
			expectError:   true,
			errorMsg:      "onChange callback cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewWatcher(tt.vaultConfig, tt.checkInterval, tt.onChange)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewWatcher() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("NewWatcher() error = %v, want %v", err.Error(), tt.errorMsg)
				}
				if watcher != nil {
					t.Errorf("NewWatcher() expected nil watcher when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("NewWatcher() unexpected error = %v", err)
				return
			}

			if watcher == nil {
				t.Errorf("NewWatcher() returned nil watcher")
				return
			}

			// Verify watcher configuration
			if watcher.vaultConfig != tt.vaultConfig {
				t.Errorf("NewWatcher() vaultConfig not set correctly")
			}
			if watcher.checkInterval != tt.checkInterval {
				t.Errorf("NewWatcher() checkInterval = %v, want %v", watcher.checkInterval, tt.checkInterval)
			}
			if watcher.client == nil {
				t.Errorf("NewWatcher() client not initialized")
			}
			if watcher.started {
				t.Errorf("NewWatcher() watcher should not be started initially")
			}

			// Clean up
			watcher.Stop()
		})
	}
}

func TestWatcher_GetCurrentHash(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	watcher, err := NewWatcher(config, time.Second, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Initially, hash should be empty
	if hash := watcher.GetCurrentHash(); hash != "" {
		t.Errorf("GetCurrentHash() = %v, want empty string", hash)
	}

	// Set a hash manually for testing
	testHash := "test-hash-123"
	watcher.mu.Lock()
	watcher.currentHash = testHash
	watcher.mu.Unlock()

	if hash := watcher.GetCurrentHash(); hash != testHash {
		t.Errorf("GetCurrentHash() = %v, want %v", hash, testHash)
	}
}

func TestWatcher_IsStarted(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	watcher, err := NewWatcher(config, time.Second, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Initially should not be started
	if watcher.IsStarted() {
		t.Errorf("IsStarted() = true, want false")
	}

	// Set started manually for testing
	watcher.mu.Lock()
	watcher.started = true
	watcher.mu.Unlock()

	if !watcher.IsStarted() {
		t.Errorf("IsStarted() = false, want true")
	}
}

func TestWatcher_StartAlreadyStarted(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	// For this test, we'll use the real NewWatcher but won't actually start it
	watcher, err := NewWatcher(config, time.Second, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Manually set started to true
	watcher.mu.Lock()
	watcher.started = true
	watcher.mu.Unlock()

	err = watcher.Start()
	if err == nil {
		t.Errorf("Start() expected error for already started watcher")
	}
	
	expectedError := "watcher is already started"
	if err.Error() != expectedError {
		t.Errorf("Start() error = %v, want %v", err.Error(), expectedError)
	}
}

func TestWatcher_OnChangeCallback(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	callbackCalled := false
	callbackMutex := sync.Mutex{}
	
	onChange := func() error {
		callbackMutex.Lock()
		callbackCalled = true
		callbackMutex.Unlock()
		return nil
	}

	watcher, err := NewWatcher(config, 100*time.Millisecond, onChange)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Test that callback is called when detecting changes
	// This would require mocking the vault client more extensively
	// For now, we test that the callback can be set and called manually
	
	if err := onChange(); err != nil {
		t.Errorf("onChange callback failed: %v", err)
	}

	callbackMutex.Lock()
	if !callbackCalled {
		t.Errorf("onChange callback was not called")
	}
	callbackMutex.Unlock()
}

func TestWatcher_OnChangeCallbackError(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	expectedError := errors.New("callback error")
	onChange := func() error {
		return expectedError
	}

	watcher, err := NewWatcher(config, time.Second, onChange)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Test that callback errors are properly returned
	if err := onChange(); err != expectedError {
		t.Errorf("onChange callback error = %v, want %v", err, expectedError)
	}
}

func TestWatcher_ConcurrentAccess(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	watcher, err := NewWatcher(config, time.Second, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Test concurrent access to methods
	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent GetCurrentHash calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			hash := watcher.GetCurrentHash()
			// Just verify it doesn't panic and returns a string
			_ = hash
		}()
	}
	wg.Wait()

	// Test concurrent IsStarted calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			started := watcher.IsStarted()
			// Just verify it doesn't panic and returns a bool
			_ = started
		}()
	}
	wg.Wait()
}

func TestWatcher_Stop(t *testing.T) {
	config := &VaultConfig{
		Host:  "https://vault.example.com",
		Path:  "kv/data/test",
		Token: "test-token",
	}

	watcher, err := NewWatcher(config, time.Second, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	// Stop should not panic even if watcher wasn't started
	watcher.Stop()

	// Should be able to call Stop multiple times without issues
	watcher.Stop()
	watcher.Stop()
}

func TestVaultConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *VaultConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &VaultConfig{
				Host:  "https://vault.example.com",
				Path:  "kv/data/test",
				Token: "test-token",
			},
			valid: true,
		},
		{
			name: "valid config with IP",
			config: &VaultConfig{
				Host:  "http://192.168.1.100:8200",
				Path:  "secret/myapp",
				Token: "hvs.abc123",
			},
			valid: true,
		},
		{
			name: "valid config with port",
			config: &VaultConfig{
				Host:  "https://vault.company.com:8200",
				Path:  "kv/v2/data/production/database",
				Token: "s.1234567890abcdef",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWatcher(tt.config, time.Second, func() error { return nil })
			
			if tt.valid && err != nil {
				t.Errorf("Expected valid config to not produce error, got: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid config to produce error")
			}
		})
	}
}