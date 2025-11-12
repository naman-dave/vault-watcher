package vaultwatcher

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// CalculateHash calculates a SHA256 hash of all variables in the vault data
func CalculateHash(vaultData map[string]interface{}) (string, error) {
	if vaultData == nil {
		return "", fmt.Errorf("vault data cannot be nil")
	}

	jsonBytes, err := json.Marshal(vaultData)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:]), nil
}
