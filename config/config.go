package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig
	Redis        RedisConfig
	Meilisearch  MeilisearchConfig
	RateLimit    RateLimitConfig
}

type ServerConfig struct {
	Port string
	APIKey string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type MeilisearchConfig struct {
	Host string
	Key  string
}

type RateLimitConfig struct {
	RequestsPerMinute int
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8076")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", "0")
	viper.SetDefault("MEILISEARCH_HOST", "http://localhost:7700")
	viper.SetDefault("RATE_LIMIT_REQUESTS", "60")

	// Read env file
	if err := viper.ReadInConfig(); err != nil {
		// Continue without .env file if not found
	}

	// Bind environment variables
	viper.AutomaticEnv()

	config := &Config{
		Server: ServerConfig{
			Port:   viper.GetString("SERVER_PORT"),
			APIKey: viper.GetString("API_KEY"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Meilisearch: MeilisearchConfig{
			Host: viper.GetString("MEILISEARCH_HOST"),
			Key:  viper.GetString("MEILISEARCH_KEY"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: viper.GetInt("RATE_LIMIT_REQUESTS"),
		},
	}

	return config, nil
}
