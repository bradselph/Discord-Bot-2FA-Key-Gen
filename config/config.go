package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken    string
	AllowedRoles    []string
	DevUserID       string
	LogLevel        string
	GuildID         string
	CommandCooldown int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		DiscordToken:    getEnv("DISCORD_BOT_TOKEN", ""),
		DevUserID:       getEnv("DEV_USER_ID", ""),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		GuildID:         getEnv("GUILD_ID", ""),
		CommandCooldown: getEnvInt("COMMAND_COOLDOWN", 5),
	}

	if roleStr := getEnv("ALLOWED_ROLES", ""); roleStr != "" {
		config.AllowedRoles = strings.Split(roleStr, ",")
		for i, role := range config.AllowedRoles {
			config.AllowedRoles[i] = strings.TrimSpace(role)
		}
	}

	if config.DiscordToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	if config.CommandCooldown < 1 {
		config.CommandCooldown = 5
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}
