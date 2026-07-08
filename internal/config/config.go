package config

import "os"

type Config struct {
	Port  string
	DbUrl string
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(key + " is required")
	}
	return v
}

func Load() *Config {
	return &Config{
		Port:  getEnv("PORT", ""),
		DbUrl: mustGetEnv("DB_URL"),
	}
}
