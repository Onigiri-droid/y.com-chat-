package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"log"
	"net/http"
)

// GetChatByIDHandler обрабатывает запрос на получение информации о чате
func GetChatByIDHandler(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Проверяем метод запроса
        if r.Method != http.MethodGet {
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        // Извлекаем chatID из URL-параметров
        chatID := r.URL.Path[len("/api/chats/"):]
        if chatID == "" {
            http.Error(w, "Отсутствует обязательный параметр chatID", http.StatusBadRequest)
            return
        }

        // Извлекаем userID из контекста (добавленного AuthMiddleware)
        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
            return
        }

        // Получаем информацию о чате
        chat, err := storage.GetChatByID(r.Context(), chatID)
        if err != nil {
            switch err.Error() {
            case "некорректный идентификатор чата":
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case "чат не найден":
                http.Error(w, "Чат не найден", http.StatusNotFound)
            default:
                log.Printf("Ошибка получения чата: %v", err)
                http.Error(w, "Не удалось получить информацию о чате", http.StatusInternalServerError)
            }
            return
        }

        // Проверяем права доступа к чату
        if !isUserAllowedToViewChat(chat, userID) {
            http.Error(w, "У вас нет прав на просмотр этого чата", http.StatusForbidden)
            return
        }

        // Возвращаем успешный ответ с информацией о чате
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(chat); err != nil {
            http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
        }
    }
}