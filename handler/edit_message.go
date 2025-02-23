package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"chat-service/middleware"
	"chat-service/storage"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EditMessageRequest struct {
	NewContent string `json:"new_content"`
}

// EditMessageHandler handles requests to edit a message.
func EditMessageHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Проверяем метод запроса
		if r.Method != http.MethodPut {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		// Извлекаем messageID из URL
		messageID := extractMessageIDFromURL(r.URL.Path)
		if !isValidMessageID(messageID) {
			http.Error(w, "Некорректный messageID", http.StatusBadRequest)
			return
		}

		// Извлекаем userID из контекста
		userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
		if !ok {
			log.Printf("Не удалось извлечь userID из контекста")
			http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
			return
		}

		// Парсим тело запроса
		var req EditMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
			return
		}

		// Проверяем наличие нового содержимого
		if strings.TrimSpace(req.NewContent) == "" {
			http.Error(w, "Новое содержимое не должно быть пустым", http.StatusBadRequest)
			return
		}

		// Вызываем метод редактирования сообщения в storage
		err := store.EditMessage(ctx, messageID, userID, req.NewContent)
		if err != nil {
			handleEditMessageError(w, err)
			return
		}

		// Возвращаем успешный ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"message": "Сообщение успешно отредактировано",
		}); err != nil {
			log.Printf("Ошибка при кодировании JSON: %v", err)
			http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
		}
	}
}

// Helper Functions

func extractMessageIDFromURL(path string) string {
	return strings.TrimPrefix(path, "/api/messages/")
}

func isValidMessageID(messageID string) bool {
	return primitive.IsValidObjectID(messageID)
}

func handleEditMessageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrMessageNotFound):
		http.Error(w, "Сообщение не найдено", http.StatusNotFound)
	case errors.Is(err, storage.ErrForbidden):
		http.Error(w, "Вы не можете редактировать чужое сообщение", http.StatusForbidden)
	default:
		log.Printf("Ошибка редактирования сообщения: %v", err)
		http.Error(w, "Не удалось отредактировать сообщение", http.StatusInternalServerError)
	}
}
