package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EventDraft представляет собой черновик события, готовый к добавлению в календарь
type EventDraft struct {
	Title string
	Start time.Time
	End   time.Time
}

// Parse разбирает сообщение пользователя и превращает его в структуру события
// Ожидается формат: "Название; дата; время; длительность"
// Пример: "Созвон; 2026-10-15; 15:00; 60"
func Parse(text string, loc *time.Location) (*EventDraft, error) {
	// Разбиваем текст по символу ";"
	parts := strings.Split(text, ";")

	// Проверяем, что есть минимум 3 части: название, дата, время
	if len(parts) < 3 {
		return nil, errors.New(
			"неверный формат. Используй: Название; дата; время; длительность в минутах\nпример: Созвон; 2026-10-15; 15:00; 60",
		)
	}

	// Извлекаем и очищаем от лишних пробелов название
	title := strings.TrimSpace(parts[0])
	if title == "" {
		return nil, errors.New("название события не может быть пустым")
	}

	// Извлекаем дату и время
	dateText := strings.TrimSpace(parts[1])
	timeText := strings.TrimSpace(parts[2])

	// Парсим дату. В Go эталонная дата для формата ГГГГ-ММ-ДД — это 2006-01-02
	date, err := time.ParseInLocation("2006-01-02", dateText, loc)
	if err != nil {
		return nil, fmt.Errorf("неверная дата. Ожидалось ГГГГ-ММ-ДД (например, 2026-10-15), получено: %q", dateText)
	}

	// Парсим время. Эталонное время для ЧЧ:ММ в Go — 15:04
	parsedTime, err := time.ParseInLocation("15:04", timeText, loc)
	if err != nil {
		return nil, fmt.Errorf("неверное время. Ожидалось ЧЧ:ММ (например, 15:00), получено: %q", timeText)
	}

	// Собираем дату и время вместе в нужном часовом поясе
	start := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0,
		0,
		loc,
	)

	// Защита от дурака: нельзя создать событие в прошлом
	if start.Before(time.Now()) {
		return nil, errors.New("событие не может быть в прошлом")
	}

	// Длительность по умолчанию: 60 минут
	duration := 60 * time.Minute

	// Если пользователь указал длительность (4-я часть), читаем её
	if len(parts) >= 4 {
		minutesText := strings.TrimSpace(parts[3])
		minutes, err := strconv.Atoi(minutesText)
		if err != nil {
			return nil, fmt.Errorf("неверная длительность. Ожидалось число минут, получено: %q", minutesText)
		}
		if minutes <= 0 {
			return nil, errors.New("длительность должна быть больше нуля")
		}
		duration = time.Duration(minutes) * time.Minute
	}

	// Вычисляем время окончания
	end := start.Add(duration)

	return &EventDraft{
		Title: title,
		Start: start,
		End:   end,
	}, nil
}
