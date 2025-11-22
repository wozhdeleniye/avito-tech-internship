package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8085"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "myuser"),
			Password: getEnv("DB_PASSWORD", "mypassword"),
			DBName:   getEnv("DB_NAME", "mydatabase"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", "access_secret"),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", "refresh_secret"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", time.Hour*24),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", time.Hour*24*7),
		},
	}
}

func (c *Config) GetServerConfig() ServerConfig {
	return c.Server
}

func (c *Config) GetDatabaseConfig() DatabaseConfig {
	return c.Database
}

func (c *Config) GetRedisConfig() RedisConfig {
	return c.Redis
}

func (c *Config) GetJWTConfig() JWTConfig {
	return c.JWT
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
