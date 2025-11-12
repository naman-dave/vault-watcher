package vaultwatcher

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR_EXISTS",
			defaultValue: "default",
			envValue:     "env_value",
			setEnv:       true,
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "TEST_VAR_NOT_EXISTS",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "environment variable exists but empty",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "empty default value",
			key:          "TEST_VAR_EMPTY_DEFAULT",
			defaultValue: "",
			envValue:     "",
			setEnv:       false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadVaultConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"VAULT_HOST":  "https://vault.example.com",
				"VAULT_PATH":  "kv/data/myapp/config",
				"VAULT_TOKEN": "test-token",
			},
			expectError: false,
		},
		{
			name: "missing VAULT_HOST",
			envVars: map[string]string{
				"VAULT_PATH":  "kv/data/myapp/config",
				"VAULT_TOKEN": "test-token",
			},
			expectError: true,
			errorMsg:    "VAULT_HOST environment variable is required",
		},
		{
			name: "missing VAULT_PATH",
			envVars: map[string]string{
				"VAULT_HOST":  "https://vault.example.com",
				"VAULT_TOKEN": "test-token",
			},
			expectError: true,
			errorMsg:    "VAULT_PATH environment variable is required",
		},
		{
			name: "missing VAULT_TOKEN",
			envVars: map[string]string{
				"VAULT_HOST": "https://vault.example.com",
				"VAULT_PATH": "kv/data/myapp/config",
			},
			expectError: true,
			errorMsg:    "VAULT_TOKEN environment variable is required",
		},
		{
			name: "empty VAULT_HOST",
			envVars: map[string]string{
				"VAULT_HOST":  "",
				"VAULT_PATH":  "kv/data/myapp/config",
				"VAULT_TOKEN": "test-token",
			},
			expectError: true,
			errorMsg:    "VAULT_HOST environment variable is required",
		},
		{
			name:        "no environment variables set",
			envVars:     map[string]string{},
			expectError: true,
			errorMsg:    "VAULT_HOST environment variable is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variables
			envKeys := []string{"VAULT_HOST", "VAULT_PATH", "VAULT_TOKEN"}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				if value != "" {
					os.Setenv(key, value)
				}
			}

			// Clean up after test
			defer func() {
				for _, key := range envKeys {
					os.Unsetenv(key)
				}
			}()

			config, err := LoadVaultConfigFromEnv()

			if tt.expectError {
				if err == nil {
					t.Errorf("LoadVaultConfigFromEnv() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("LoadVaultConfigFromEnv() error = %v, want %v", err.Error(), tt.errorMsg)
				}
				if config != nil {
					t.Errorf("LoadVaultConfigFromEnv() expected nil config when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadVaultConfigFromEnv() unexpected error = %v", err)
				return
			}

			if config == nil {
				t.Errorf("LoadVaultConfigFromEnv() returned nil config")
				return
			}

			// Verify config values
			if config.Host != tt.envVars["VAULT_HOST"] {
				t.Errorf("LoadVaultConfigFromEnv() Host = %v, want %v", config.Host, tt.envVars["VAULT_HOST"])
			}
			if config.Path != tt.envVars["VAULT_PATH"] {
				t.Errorf("LoadVaultConfigFromEnv() Path = %v, want %v", config.Path, tt.envVars["VAULT_PATH"])
			}
			if config.Token != tt.envVars["VAULT_TOKEN"] {
				t.Errorf("LoadVaultConfigFromEnv() Token = %v, want %v", config.Token, tt.envVars["VAULT_TOKEN"])
			}
		})
	}
}
