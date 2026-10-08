package tgbot

import (
	"context"
	"fmt"
	"log"

	"tgbot-calendar/internal/config"
	"tgbot-calendar/internal/gcal"
	"tgbot-calendar/internal/parser"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RunBot запускает Telegram бота и обрабатывает входящие сообщения
func RunBot(ctx context.Context, cfg *config.Config, calendarService *gcal.Service) error {
	// Создаем экземпляр бота с токеном
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return fmt.Errorf("ошибка при создании бота: %w", err)
	}

	log.Printf("✅ Бот авторизован как @%s", bot.Self.UserName)

	// Настройка получения обновлений (сообщений от пользователей)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("👀 Ожидаю входящих сообщений...")

	// Бесконечный цикл обработки сообщений
	for update := range updates {
		// Игнорируем обновления без сообщений (например, нажатия кнопок пока не обрабатываем)
		if update.Message == nil {
			continue
		}

		// Игнорируем сообщения без текста
		if update.Message.Text == "" {
			continue
		}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		// 🔒 ПРОВЕРКА ДОСТУПА
		// Если пользователя нет в списке разрешенных, игнорируем его
		if !cfg.IsAllowed(userID) {
			log.Printf("🚫 Попытка доступа от неразрешенного пользователя (ID: %d)", userID)
			sendMessage(bot, chatID, "🚫 Доступ запрещен.")
			continue
		}

		// 📖 ОБРАБОТКА КОМАНД
		if text == "/start" || text == "/help" {
			helpText := "Привет! Я бот для добавления событий в твой календарь. 📅\n\n" +
				"Отправь мне сообщение в формате:\n" +
				"`Название; дата; время; длительность в минутах`\n\n" +
				"Пример:\n" +
				"`Созвон; 2026-10-15; 15:00; 60`\n" +
				"Если длительность не указать, по умолчанию будет 60 минут."
			sendMessage(bot, chatID, helpText)
			continue
		}

		// 🧠 ПАРСИНГ СООБЩЕНИЯ
		draft, err := parser.Parse(text, cfg.Timezone)
		if err != nil {
			sendMessage(bot, chatID, fmt.Sprintf("❌ Ошибка парсинга:\n%v", err))
			continue
		}

		// 📅 СОЗДАНИЕ СОБЫТИЯ В КАЛЕНДАРЕ
		createdEvent, err := calendarService.CreateEvent(ctx, draft)
		if err != nil {
			sendMessage(bot, chatID, fmt.Sprintf("❌ Ошибка при создании события в календаре:\n%v", err))
			continue
		}

		// ✅ УСПЕХ
		response := fmt.Sprintf("✅ Событие добавлено!\n\n"+
			"📌 Название: %s\n"+
			"🕒 Начало: %s (%s)\n"+
			"🔚 Конец: %s (%s)\n\n"+
			"🔗 Ссылка: %s",
			draft.Title,
			draft.Start.Format("2006-01-02 15:04"), draft.Start.Location(),
			draft.End.Format("2006-01-02 15:04"), draft.End.Location(),
			createdEvent.HtmlLink,
		)
		sendMessage(bot, chatID, response)
	}

	return nil
}

// sendMessage отправляет текстовое сообщение в чат
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
