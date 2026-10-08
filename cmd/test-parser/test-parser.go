package main

import (
	"fmt"
	"log"

	// ЗАМЕНИ "tgbot-calendar" на имя из твоего go.mod
	"tgbot-calendar/internal/config"
	"tgbot-calendar/internal/parser"
)

func main() {
	// Загружаем настройки, чтобы взять часовой пояс
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	// Список тестовых сообщений
	testMessages := []string{
		"Созвон с командой; 2026-12-31; 15:00; 60", // Идеальный вариант
		"Обед; 2026-06-15; 13:30",                  // Без длительности (по умолчанию 60 мин)
		"Неверный формат без даты",                 // Должен выдать ошибку
		"Встреча в прошлом; 2020-01-01; 10:00; 30", // Должен выдать ошибку (прошлое)
	}

	fmt.Println("🧪 Тестирование парсера событий:")
	fmt.Println("==============================================")

	for _, msg := range testMessages {
		draft, err := parser.Parse(msg, cfg.Timezone)

		if err != nil {
			// Если парсер выдал ошибку
			fmt.Printf("❌ ОШИБКА для \"%s\"\n", msg)
			fmt.Printf("   Причина: %v\n\n", err)
			continue
		}

		// Если парсер сработал успешно
		fmt.Printf("✅ УСПЕХ для \"%s\"\n", msg)
		fmt.Printf("   Название: %s\n", draft.Title)
		fmt.Printf("   Начало:   %s (%s)\n", draft.Start.Format("2006-01-02 15:04"), draft.Start.Location())
		fmt.Printf("   Конец:    %s (%s)\n\n", draft.End.Format("2006-01-02 15:04"), draft.End.Location())
	}
}
