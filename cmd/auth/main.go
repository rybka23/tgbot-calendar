package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	// 1. Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, читаем переменные из окружения...")
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	tokenFile := os.Getenv("TOKEN_FILE")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		log.Fatal("Ошибка: Не найдены GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET или GOOGLE_REDIRECT_URL в .env")
	}
	if tokenFile == "" {
		tokenFile = "token.json"
	}

	// 2. Настраиваем OAuth конфиг
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.events"},
		Endpoint:     google.Endpoint,
	}

	// 3. Генерируем случайный state для безопасности
	state, err := generateState()
	if err != nil {
		log.Fatalf("Не удалось сгенерировать state: %v", err)
	}

	// 4. Настраиваем локальный сервер для приема callback от Google
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем state
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Неверный state", http.StatusBadRequest)
			errCh <- fmt.Errorf("неверный state от Google")
			return
		}

		// Забираем код
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Код не найден", http.StatusBadRequest)
			errCh <- fmt.Errorf("Google не вернул код")
			return
		}

		// Отвечаем браузеру, что всё ок
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<h1>Авторизация успешна!</h1><p>Можешь закрыть эту вкладку и вернуться в терминал.</p>")

		codeCh <- code
	})

	server := &http.Server{Addr: "127.0.0.1:8085", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("ошибка локального сервера: %v", err)
		}
	}()

	// 5. Формируем ссылку для браузера
	// oauth2.AccessTypeOffline - чтобы дали refresh_token
	// prompt=consent - принудительно показать экран выбора, чтобы refresh_token точно выдали
	authURL := conf.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	fmt.Println("\n🔗 Открой эту ссылку в браузере:")
	fmt.Println(authURL)
	fmt.Println("\nОжидание ответа от Google...")

	// 6. Ждем код или ошибку
	var code string
	select {
	case code = <-codeCh:
		// Успех
	case err := <-errCh:
		log.Fatalf("Ошибка при получении кода: %v", err)
	case <-time.After(3 * time.Minute):
		log.Fatal("Таймаут: ты не успел авторизоваться за 3 минуты")
	}

	// Останавливаем сервер, он больше не нужен
	server.Shutdown(context.Background())

	// 7. Меняем код на токены
	ctx := context.Background()
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Не удалось обменять код на токен: %v", err)
	}

	// 8. Проверяем refresh_token
	if token.RefreshToken == "" {
		log.Println("⚠️ ВНИМАНИЕ: refresh_token пустой!")
		log.Println("Это значит, что Google не дал долгосрочный доступ.")
		log.Println("Зайди в https://myaccount.google.com/permissions, удали доступ этому приложению и запусти скрипт заново.")
	}

	// 9. Сохраняем в файл
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		log.Fatalf("Не удалось сериализовать токен: %v", err)
	}

	err = os.WriteFile(tokenFile, data, 0600) // 0600 - права только для владельца
	if err != nil {
		log.Fatalf("Не удалось сохранить файл %s: %v", tokenFile, err)
	}

	fmt.Printf("\n✅ Успех! Токен сохранен в файл: %s\n", tokenFile)
}

// Вспомогательная функция для генерации случайной строки (state)
func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
