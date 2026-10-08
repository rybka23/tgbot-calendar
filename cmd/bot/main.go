package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"tgbot-calendar/internal/config"
	"tgbot-calendar/internal/gcal"
	"tgbot-calendar/internal/tgbot"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	// Инициализируем контекст для корректного завершения по сигналу (например, Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Создаем сервис календаря (используем token.json)
	calendarService, err := gcal.NewService(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации календаря: %v", err)
	}

	log.Println("🚀 Запуск Telegram бота...")

	// Запускаем бота
	if err := tgbot.RunBot(ctx, cfg, calendarService); err != nil {
		log.Fatalf("❌ Ошибка работы бота: %v", err)
	}
}
