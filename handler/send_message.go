package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"log"
	"net/http"
)

// SendMessageHandler обрабатывает отправку сообщения в чат
func SendMessageHandler(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Обработка запроса на отправку сообщения")

        // Проверяем метод запроса
        if r.Method != http.MethodPost {
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        // Извлекаем senderID из контекста (добавленного AuthMiddleware)
        senderID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь senderID из контекста")
            http.Error(w, "Не удалось извлечь senderID из токена", http.StatusInternalServerError)
            return
        }

        // Читаем тело запроса
        var req struct {
            ChatID     string `json:"chat_id"`
            Content    string `json:"content"`
            MessageType string `json:"type"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Некорректный запрос", http.StatusBadRequest)
            return
        }

        // Проверяем обязательные поля
        if req.ChatID == "" || req.Content == "" || req.MessageType == "" {
            http.Error(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
            return
        }

        // Получаем информацию о чате
        chat, err := storage.GetChatByID(r.Context(), req.ChatID)
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

        // Проверяем права пользователя на отправку сообщения
        if !isUserAllowedToSendMessage(chat, senderID) {
            http.Error(w, "У вас нет прав на отправку сообщений в этот чат", http.StatusForbidden)
            return
        }

        // Сохраняем новое сообщение
        messageID, err := storage.SaveMessage(r.Context(), req.ChatID, senderID, req.Content, req.MessageType)
        if err != nil {
            switch err.Error() {
            case "некорректный идентификатор чата":
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            default:
                log.Printf("Ошибка при сохранении сообщения: %v", err)
                http.Error(w, "Не удалось отправить сообщение", http.StatusInternalServerError)
            }
            return
        }

        // Возвращаем успешный ответ с ID созданного сообщения
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        if err := json.NewEncoder(w).Encode(map[string]string{"message_id": messageID}); err != nil {
            http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
        }
    }
}

// isUserAllowedToSendMessage проверяет, имеет ли пользователь права на отправку сообщений в чат
func isUserAllowedToSendMessage(chat *storage.Chat, userID int32) bool {
    // Проверяем, является ли пользователь участником чата
    for _, memberID := range chat.MemberIDs {
        if memberID == userID {
            return true
        }
    }
    return false
}