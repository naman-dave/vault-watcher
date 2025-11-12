package vaultwatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultConfig holds the Vault connection configuration
type VaultConfig struct {
	Host  string // VAULT_HOST
	Path  string // VAULT_PATH
	Token string // VAULT_TOKEN
}

// Watcher monitors a Vault path for changes by comparing hashes of the variables
type Watcher struct {
	vaultConfig   *VaultConfig
	client        *api.Client
	currentHash   string
	checkInterval time.Duration
	onChange      func() error
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	started       bool
}

// NewWatcher creates a new Vault watcher instance
// vaultConfig: Vault connection configuration
// checkInterval: How often to check for changes (e.g., 30 * time.Second)
// onChange: Callback function to execute when changes are detected
func NewWatcher(vaultConfig *VaultConfig, checkInterval time.Duration, onChange func() error) (*Watcher, error) {
	if vaultConfig == nil {
		return nil, fmt.Errorf("vault config cannot be nil")
	}
	if vaultConfig.Host == "" {
		return nil, fmt.Errorf("VAULT_HOST is required")
	}
	if vaultConfig.Path == "" {
		return nil, fmt.Errorf("VAULT_PATH is required")
	}
	if vaultConfig.Token == "" {
		return nil, fmt.Errorf("VAULT_TOKEN is required")
	}
	if onChange == nil {
		return nil, fmt.Errorf("onChange callback cannot be nil")
	}

	// Create Vault client
	vaultClientConfig := api.DefaultConfig()
	vaultClientConfig.Address = vaultConfig.Host

	client, err := api.NewClient(vaultClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set the token
	client.SetToken(vaultConfig.Token)

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		vaultConfig:   vaultConfig,
		client:        client,
		checkInterval: checkInterval,
		onChange:      onChange,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// LoadVaultConfigFromEnv loads Vault connection details from environment variables
func LoadVaultConfigFromEnv() (*VaultConfig, error) {
	host := getEnv("VAULT_HOST", "")
	path := getEnv("VAULT_PATH", "")
	token := getEnv("VAULT_TOKEN", "")

	if host == "" {
		return nil, fmt.Errorf("VAULT_HOST environment variable is required")
	}
	if path == "" {
		return nil, fmt.Errorf("VAULT_PATH environment variable is required")
	}
	if token == "" {
		return nil, fmt.Errorf("VAULT_TOKEN environment variable is required")
	}

	return &VaultConfig{
		Host:  host,
		Path:  path,
		Token: token,
	}, nil
}

// fetchVaultData reads data from Vault and returns it as a map
func (w *Watcher) fetchVaultData() (map[string]interface{}, error) {
	// Read secret from Vault
	secret, err := w.client.Logical().Read(w.vaultConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("failed to read secret from vault: secret is nil")
	}
	if secret.Data == nil {
		return nil, fmt.Errorf("failed to read secret from vault: secret data is nil")
	}

	var vaultData map[string]interface{}
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		// KV v2 format
		vaultData = data
	} else {
		// KV v1 format or direct data
		vaultData = secret.Data
	}

	return vaultData, nil
}

// Start begins monitoring the Vault path for changes
// It calculates the initial hash and then periodically checks for changes
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return fmt.Errorf("watcher is already started")
	}
	w.started = true
	w.mu.Unlock()

	// Calculate initial hash
	vaultData, err := w.fetchVaultData()
	if err != nil {
		return fmt.Errorf("failed to fetch initial vault data: %w", err)
	}

	initialHash, err := CalculateHash(vaultData)
	if err != nil {
		return fmt.Errorf("failed to calculate initial hash: %w", err)
	}

	w.mu.Lock()
	w.currentHash = initialHash
	w.mu.Unlock()

	// Start the monitoring goroutine
	w.wg.Add(1)
	go w.monitor()

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.cancel()
	w.wg.Wait()

	w.mu.Lock()
	w.started = false
	w.mu.Unlock()
}

// monitor runs in a goroutine and periodically checks for changes
func (w *Watcher) monitor() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.checkForChanges(); err != nil {
				// Log error but continue monitoring
				// You might want to add a logger here
				fmt.Printf("Error checking for vault changes: %v\n", err)
				continue
			}
		}
	}
}

// checkForChanges fetches the current vault data, calculates its hash,
// and compares it with the stored hash. If different, calls the onChange callback.
func (w *Watcher) checkForChanges() error {
	vaultData, err := w.fetchVaultData()
	if err != nil {
		return fmt.Errorf("failed to fetch vault data: %w", err)
	}

	newHash, err := CalculateHash(vaultData)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	w.mu.RLock()
	currentHash := w.currentHash
	w.mu.RUnlock()

	if newHash != currentHash {
		// Hash changed, execute callback
		if err := w.onChange(); err != nil {
			return fmt.Errorf("onChange callback failed: %w", err)
		}

		// Update the current hash
		w.mu.Lock()
		w.currentHash = newHash
		w.mu.Unlock()
	}

	return nil
}

// GetCurrentHash returns the current hash of the vault data
func (w *Watcher) GetCurrentHash() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentHash
}

// IsStarted returns whether the watcher is currently running
func (w *Watcher) IsStarted() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.started
}
