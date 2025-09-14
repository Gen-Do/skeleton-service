package env

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns default when env var not set",
			key:          "TEST_STRING_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_STRING_SET",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_STRING_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			os.Unsetenv(tt.key)

			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Test
			result := GetString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns default when env var not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "returns default for invalid value",
			key:          "TEST_BOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			os.Unsetenv(tt.key)

			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Test
			result := GetBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float64
		envValue     string
		expected     float64
	}{
		{
			name:         "returns default when env var not set",
			key:          "TEST_FLOAT_NOT_SET",
			defaultValue: 1.5,
			envValue:     "",
			expected:     1.5,
		},
		{
			name:         "returns parsed float",
			key:          "TEST_FLOAT_SET",
			defaultValue: 1.0,
			envValue:     "2.5",
			expected:     2.5,
		},
		{
			name:         "returns default for invalid value",
			key:          "TEST_FLOAT_INVALID",
			defaultValue: 1.0,
			envValue:     "not_a_float",
			expected:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			os.Unsetenv(tt.key)

			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			// Test
			result := GetFloat64(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetFloat64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMustGetString(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		key := "TEST_MUST_GET_SET"
		expected := "test_value"

		os.Setenv(key, expected)
		defer os.Unsetenv(key)

		result := MustGetString(key)
		if result != expected {
			t.Errorf("MustGetString() = %v, want %v", result, expected)
		}
	})

	t.Run("panics when env var is not set", func(t *testing.T) {
		key := "TEST_MUST_GET_NOT_SET"
		os.Unsetenv(key)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustGetString() did not panic")
			}
		}()

		MustGetString(key)
	})
}

func TestIsSet(t *testing.T) {
	t.Run("returns true when env var is set", func(t *testing.T) {
		key := "TEST_IS_SET_TRUE"
		os.Setenv(key, "value")
		defer os.Unsetenv(key)

		if !IsSet(key) {
			t.Errorf("IsSet() = false, want true")
		}
	})

	t.Run("returns false when env var is not set", func(t *testing.T) {
		key := "TEST_IS_SET_FALSE"
		os.Unsetenv(key)

		if IsSet(key) {
			t.Errorf("IsSet() = true, want false")
		}
	})

	t.Run("returns true when env var is set to empty", func(t *testing.T) {
		key := "TEST_IS_SET_EMPTY"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		if !IsSet(key) {
			t.Errorf("IsSet() = false, want true")
		}
	})
}

func TestLoadEnvFiles(t *testing.T) {
	t.Run("loads env files without error when files don't exist", func(t *testing.T) {
		// Функция не должна паниковать или возвращать ошибку
		LoadEnvFiles()
	})
}
