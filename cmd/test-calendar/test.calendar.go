package main

import (
	"context"
	"fmt"
	"log"
	"tgbot-calendar/internal/config"
	"tgbot-calendar/internal/gcal"
	"tgbot-calendar/internal/parser"
	"time"
)

func main() {
	// 1. Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка конфига: %v", err)
	}

	// 2. Создаем тестовое событие на завтрашний день
	tomorrow := time.Now().Add(24 * time.Hour)

	draft := &parser.EventDraft{
		Title: "Тестовое событие от Go-бота",
		Start: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, cfg.Timezone),
		End:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 13, 0, 0, 0, cfg.Timezone),
	}

	fmt.Println("🗓️ Пытаюсь создать событие в календаре...")
	fmt.Printf("   Название: %s\n", draft.Title)
	fmt.Printf("   Время: %s - %s\n\n", draft.Start.Format("2006-01-02 15:04"), draft.End.Format("2006-01-02 15:04"))

	// 3. Создаем сервис календаря
	ctx := context.Background()
	gcalService, err := gcal.NewService(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Не удалось инициализировать календарь: %v", err)
	}

	// 4. Отправляем событие в Google
	createdEvent, err := gcalService.CreateEvent(ctx, draft)
	if err != nil {
		log.Fatalf("❌ Ошибка создания события: %v", err)
	}

	// 5. Успех!
	fmt.Println("✅ Событие успешно создано!")
	fmt.Printf("🔗 Ссылка на событие в календаре:\n%s\n", createdEvent.HtmlLink)
	fmt.Println("\nОткрой ссылку или зайди в календарь, чтобы проверить!")
}
