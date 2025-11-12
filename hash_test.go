package vaultwatcher

import (
	"testing"
)

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name: "simple string values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "mixed data types",
			input: map[string]interface{}{
				"string":  "value",
				"int":     42,
				"float":   3.14,
				"bool":    true,
				"array":   []interface{}{"a", "b", "c"},
				"nested":  map[string]interface{}{"inner": "value"},
			},
			wantErr: false,
		},
		{
			name: "empty map",
			input: map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "vault data cannot be nil",
		},
		{
			name: "complex nested structure",
			input: map[string]interface{}{
				"database": map[string]interface{}{
					"host":     "localhost",
					"port":     5432,
					"username": "admin",
					"password": "secret",
					"ssl":      true,
				},
				"redis": map[string]interface{}{
					"host": "redis-server",
					"port": 6379,
				},
				"features": []interface{}{
					"feature1",
					"feature2",
					map[string]interface{}{
						"name":    "feature3",
						"enabled": true,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := CalculateHash(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateHash() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("CalculateHash() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateHash() unexpected error = %v", err)
				return
			}

			if hash == "" {
				t.Errorf("CalculateHash() returned empty hash")
			}

			// Hash should be 64 characters (SHA256 hex)
			if len(hash) != 64 {
				t.Errorf("CalculateHash() hash length = %d, want 64", len(hash))
			}
		})
	}
}

func TestCalculateHashConsistency(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 42,
	}

	hash1, err := CalculateHash(input)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	hash2, err := CalculateHash(input)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("CalculateHash() should be consistent: hash1=%s, hash2=%s", hash1, hash2)
	}
}

func TestCalculateHashDifferentOrder(t *testing.T) {
	// Test that hash is the same regardless of map iteration order
	input1 := map[string]interface{}{
		"a": "value1",
		"b": "value2",
		"c": "value3",
	}

	input2 := map[string]interface{}{
		"c": "value3",
		"a": "value1",
		"b": "value2",
	}

	hash1, err := CalculateHash(input1)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	hash2, err := CalculateHash(input2)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("CalculateHash() should be same for same data regardless of order: hash1=%s, hash2=%s", hash1, hash2)
	}
}

func TestCalculateHashDetectsChanges(t *testing.T) {
	originalData := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	originalHash, err := CalculateHash(originalData)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	// Test value change
	modifiedData := map[string]interface{}{
		"key1": "changed_value",
		"key2": "value2",
	}

	modifiedHash, err := CalculateHash(modifiedData)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if originalHash == modifiedHash {
		t.Error("CalculateHash() should detect value changes")
	}

	// Test new key addition
	addedData := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "new_value",
	}

	addedHash, err := CalculateHash(addedData)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if originalHash == addedHash {
		t.Error("CalculateHash() should detect new key additions")
	}

	// Test key removal
	removedData := map[string]interface{}{
		"key1": "value1",
	}

	removedHash, err := CalculateHash(removedData)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if originalHash == removedHash {
		t.Error("CalculateHash() should detect key removals")
	}
}