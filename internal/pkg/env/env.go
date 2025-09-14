package env

import (
	"os"
	"strconv"
)

// GetString возвращает строковое значение переменной окружения или значение по умолчанию
func GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetBool возвращает булево значение переменной окружения или значение по умолчанию
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// GetInt возвращает целочисленное значение переменной окружения или значение по умолчанию
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetInt64 возвращает 64-битное целочисленное значение переменной окружения или значение по умолчанию
func GetInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetFloat64 возвращает значение с плавающей точкой переменной окружения или значение по умолчанию
func GetFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// MustGetString возвращает строковое значение переменной окружения или паникует, если она не установлена
func MustGetString(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic("required environment variable " + key + " is not set")
}

// IsSet проверяет, установлена ли переменная окружения
func IsSet(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}
