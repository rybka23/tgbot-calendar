package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения в удобном виде
type Config struct {
	// Telegram
	BotToken    string
	BotUsername string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	TokenFile          string

	// App Settings
	Timezone       *time.Location
	AllowedUserIDs map[int64]bool
}

// IsAllowed проверяет, есть ли у пользователя доступ к боту
func (c *Config) IsAllowed(userID int64) bool {
	// Если список пуст, считаем, что доступ запрещен всем (безопасность)
	if len(c.AllowedUserIDs) == 0 {
		return false
	}
	return c.AllowedUserIDs[userID]
}

// Load читает .env и формирует объект Config
func Load() (*Config, error) {
	// Пытаемся загрузить .env. Если его нет, берем из системного окружения (нормально для продакшена)
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ Файл .env не найден, читаем переменные из системного окружения...")
	}

	cfg := &Config{
		BotToken:           os.Getenv("BOT_TOKEN"),
		BotUsername:        os.Getenv("BOT_USERNAME"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		TokenFile:          os.Getenv("TOKEN_FILE"),
		AllowedUserIDs:     make(map[int64]bool),
	}

	// Если TOKEN_FILE не задан, ставим дефолтный
	if cfg.TokenFile == "" {
		cfg.TokenFile = "token.json"
	}

	// --- Парсим TIMEZONE ---
	tz := os.Getenv("TIMEZONE")
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("неверный часовой пояс %q в TIMEZONE: %w", tz, err)
	}
	cfg.Timezone = loc

	// --- Парсим ALLOWED_USER_IDS ---
	// Поддерживает форматы: "123" или "123,456, 789"
	idsStr := os.Getenv("ALLOWED_USER_IDS")
	if idsStr != "" {
		for _, idPart := range strings.Split(idsStr, ",") {
			idPart = strings.TrimSpace(idPart)
			if idPart == "" {
				continue
			}
			id, err := strconv.ParseInt(idPart, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("неверный ID %q в ALLOWED_USER_IDS: %w", idPart, err)
			}
			cfg.AllowedUserIDs[id] = true
		}
	}

	return cfg, nil
}
