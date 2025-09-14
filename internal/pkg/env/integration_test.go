package env

import (
	"os"
	"testing"
)

// TestLoadEnvFilesIntegration демонстрирует работу загрузки и переопределения переменных
func TestLoadEnvFilesIntegration(t *testing.T) {
	// Сохраняем исходные значения переменных окружения
	originalServiceName := os.Getenv("SERVICE_NAME")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Очищаем переменные перед тестом
	os.Unsetenv("SERVICE_NAME")
	os.Unsetenv("LOG_LEVEL")

	// Восстанавливаем после теста
	defer func() {
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("LOG_LEVEL")
		if originalServiceName != "" {
			os.Setenv("SERVICE_NAME", originalServiceName)
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		}
	}()

	// Создаем временные файлы
	paasContent := `SERVICE_NAME=paas-service
LOG_LEVEL=info
PAAS_ONLY_VAR=paas_value`

	overrideContent := `SERVICE_NAME=override-service
OVERRIDE_ONLY_VAR=override_value`

	// Записываем временные файлы
	if err := os.WriteFile(".env.paas", []byte(paasContent), 0644); err != nil {
		t.Fatalf("Failed to create .env.paas: %v", err)
	}
	defer os.Remove(".env.paas")

	if err := os.WriteFile(".env.override", []byte(overrideContent), 0644); err != nil {
		t.Fatalf("Failed to create .env.override: %v", err)
	}
	defer os.Remove(".env.override")

	// Загружаем переменные
	LoadEnvFiles()

	// Проверяем результаты
	tests := []struct {
		name     string
		key      string
		expected string
		reason   string
	}{
		{
			name:     "SERVICE_NAME should be overridden",
			key:      "SERVICE_NAME",
			expected: "override-service",
			reason:   ".env.override should override .env.paas",
		},
		{
			name:     "LOG_LEVEL should remain from paas",
			key:      "LOG_LEVEL",
			expected: "info",
			reason:   ".env.paas value should be kept when not overridden",
		},
		{
			name:     "PAAS_ONLY_VAR should be available",
			key:      "PAAS_ONLY_VAR",
			expected: "paas_value",
			reason:   "Variables only in .env.paas should be available",
		},
		{
			name:     "OVERRIDE_ONLY_VAR should be available",
			key:      "OVERRIDE_ONLY_VAR",
			expected: "override_value",
			reason:   "Variables only in .env.override should be available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := os.Getenv(tt.key)
			if actual != tt.expected {
				t.Errorf("Expected %s=%s, got %s. Reason: %s", tt.key, tt.expected, actual, tt.reason)
			}
		})
	}
}
