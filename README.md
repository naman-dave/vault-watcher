# Vault Watcher

A reusable Go library for monitoring HashiCorp Vault paths for changes. The library uses hash-based comparison to efficiently detect when variables in Vault have changed, including when new variables are added or existing ones are modified.

## Features

- **Hash-based change detection**: Uses SHA256 hashing to detect any changes in Vault variables
- **Automatic new variable detection**: Detects when new variables are added to Vault
- **Efficient comparison**: Only compares hashes instead of full data structures
- **Callback mechanism**: Execute custom logic when changes are detected
- **Thread-safe**: Safe for concurrent use
- **Configurable polling interval**: Set how often to check for changes

## Installation


1. Add this repository as a dependency in your `go.mod`:
```bash
go get github.com/naman-dave/vault-watcher
```

Or if using a different module path, adjust accordingly.

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "os"
    "syscall"
    "time"
    
    "github.com/naman-dave/vault-watcher"
)

func main() {
    // Load Vault config from environment variables
    vaultConfig, err := vaultwatcher.LoadVaultConfigFromEnv()
    if err != nil {
        panic(err)
    }
    
    // Define what to do when changes are detected
    onChange := func() error {
        fmt.Println("Vault configuration changed!")
        // Your custom logic here, e.g., reload config, restart service, etc.
        process, _ := os.FindProcess(os.Getpid())
        process.Signal(syscall.SIGTERM)
        return nil
    }
    
    // Create watcher (checks every 30 seconds)
    watcher, err := vaultwatcher.NewWatcher(
        vaultConfig,
        30*time.Second,
        onChange,
    )
    if err != nil {
        panic(err)
    }
    
    // Start monitoring
    if err := watcher.Start(); err != nil {
        panic(err)
    }
    
    // Keep the program running
    select {}
}
```

### Manual Vault Config

```go
vaultConfig := &vaultwatcher.VaultConfig{
    Host:  "https://vault.example.com",
    Path:  "kv/data/myapp/config",
    Token: "your-vault-token",
}

watcher, err := vaultwatcher.NewWatcher(vaultConfig, 30*time.Second, onChange)
```

### Stopping the Watcher

```go
// Stop the watcher when done
watcher.Stop()
```

### Getting Current Hash

```go
currentHash := watcher.GetCurrentHash()
fmt.Printf("Current hash: %s\n", currentHash)
```

## Environment Variables

When using `LoadVaultConfigFromEnv()`, the following environment variables are required:

- `VAULT_HOST`: The Vault server address (e.g., `https://vault.example.com`)
- `VAULT_PATH`: The path to the secret in Vault (e.g., `kv/data/myapp/config`)
- `VAULT_TOKEN`: The Vault authentication token

## How It Works

1. **Initial Hash Calculation**: When the watcher starts, it fetches all variables from the specified Vault path and calculates a SHA256 hash.

2. **Periodic Checking**: At the configured interval, the watcher:
   - Fetches the current variables from Vault
   - Calculates a new hash
   - Compares it with the stored hash

3. **Change Detection**: If the hashes differ, the `onChange` callback is executed. The hash comparison ensures that:
   - Modified values are detected
   - New variables are detected
   - Removed variables are detected

4. **Hash Calculation**: The hash is calculated by:
   - Sorting all variable keys alphabetically
   - Concatenating `key:value` pairs in sorted order
   - Computing SHA256 of the concatenated string

This approach ensures deterministic hashing and efficient comparison.

## Testing

This package includes comprehensive unit tests and integration tests.

### Running Unit Tests

```bash
# Run all unit tests
go test -v

# Run tests with coverage
go test -v -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Running Integration Tests

Integration tests require a running Vault instance:

```bash
# Start Vault dev server
vault server -dev

# Set environment variables
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_HOST='http://127.0.0.1:8200'
export VAULT_PATH='secret/test'
export VAULT_TOKEN='<your-dev-token>'

# Create test secret
vault kv put secret/test key1=value1 key2=value2

# Run integration tests
go test -tags=integration -v

# To test change detection, modify the secret while tests are running:
vault kv patch secret/test key1=modified_value
```

### Test Coverage

The test suite covers:
- ✅ Hash calculation with various data types
- ✅ Hash consistency and change detection
- ✅ Environment variable loading
- ✅ Watcher creation and validation
- ✅ Concurrent access safety
- ✅ Error handling scenarios
- ✅ Integration with real Vault instances

## Thread Safety

The `Watcher` struct is thread-safe and can be used concurrently. All internal state is protected by mutexes.

## Error Handling

The watcher continues monitoring even if individual checks fail. Errors during change detection are logged but don't stop the watcher. If the `onChange` callback returns an error, it's logged but monitoring continues.

## Notes

- The watcher supports both KV v1 and KV v2 Vault secret engines
- The hash calculation handles various data types (strings, numbers, booleans, arrays, nested maps)
- The watcher runs in a separate goroutine and doesn't block the main thread

