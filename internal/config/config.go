package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port            string
	MaxUploadSize   int64
	ExportDirectory string
	UploadDirectory string
	TemplateDir     string
	StaticDir       string
	Environment     string
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	return &Config{
		Port:            getEnv("PORT", "3000"),
		MaxUploadSize:   getEnvAsInt64("MAX_UPLOAD_SIZE", 1024*1024*1024), // 1GB default
		ExportDirectory: getEnv("EXPORT_DIR", "./exports"),
		UploadDirectory: getEnv("UPLOAD_DIR", "./uploads"),
		TemplateDir:     getEnv("TEMPLATE_DIR", "./internal/templates"),
		StaticDir:       getEnv("STATIC_DIR", "./static"),
		Environment:     getEnv("ENVIRONMENT", "development"),
	}
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Helper function to get an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get an environment variable as an integer or a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get an environment variable as an int64 or a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get an environment variable as a boolean or a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
