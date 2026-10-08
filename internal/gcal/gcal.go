package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"tgbot-calendar/internal/config"
	"tgbot-calendar/internal/parser"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service отвечает за работу с Google Calendar
type Service struct {
	calendar *calendar.Service
}

// NewService создает клиент для работы с календарём, используя токен из файла
func NewService(ctx context.Context, cfg *config.Config) (*Service, error) {
	// 1. Читаем файл с токеном (который мы получили на Этапе 3)
	data, err := os.ReadFile(cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл токена %q: %w", cfg.TokenFile, err)
	}

	// 2. Расшифровываем JSON в структуру токена
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("не удалось разобрать файл токена: %w", err)
	}

	// 3. Настраиваем OAuth конфигурацию (такую же, как в auth.go)
	conf := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.events"},
		Endpoint:     google.Endpoint,
	}

	// 4. Создаем источник токенов. Если access_token истек, он сам использует
	// refresh_token для получения нового. Тебе не нужно делать это вручную.
	tokenSource := conf.TokenSource(ctx, &token)

	// 5. Создаем клиент календаря
	calendarService, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиент календаря: %w", err)
	}

	return &Service{calendar: calendarService}, nil
}

// CreateEvent добавляет событие в основной календарь пользователя
func (s *Service) CreateEvent(ctx context.Context, draft *parser.EventDraft) (*calendar.Event, error) {
	event := &calendar.Event{
		Summary: draft.Title,
		Start: &calendar.EventDateTime{
			DateTime: draft.Start.Format(time.RFC3339),
			TimeZone: draft.Start.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: draft.End.Format(time.RFC3339),
			TimeZone: draft.End.Location().String(),
		},
	}

	// Вставляем событие в "primary" (основной) календарь
	createdEvent, err := s.calendar.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании события в календаре: %w", err)
	}

	return createdEvent, nil
}
