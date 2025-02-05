package handler

import (
	"encoding/json"
	"net/http"
	"chat-service/storage"
)

// SendMessageRequest представляет данные запроса для отправки сообщения
type SendMessageRequest struct {
	ChatID  string `json:"chatID"`  // ID чата, в который отправляется сообщение
	Content string `json:"content"` // Содержимое сообщения
	Type    string `json:"type"`    // Тип сообщения (например, "text", "image", "file")
}

// SendMessageHandler обрабатывает отправку сообщения в чат
func SendMessageHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем тело запроса
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что все обязательные поля переданы
	if req.ChatID == "" || req.Content == "" {
		http.Error(w, "Не указаны chatID или content", http.StatusBadRequest)
		return
	}

	// Если тип сообщения не указан, используем значение по умолчанию "text"
	if req.Type == "" {
		req.Type = "text"
	}

	// Вызываем метод хранилища для сохранения сообщения
	messageID, err := storage.SaveMessage(r.Context(), req.ChatID, 0, req.Content, req.Type) // senderID пока не используется
	if err != nil {
		http.Error(w, "Не удалось отправить сообщение: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ с ID сообщения
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"messageID": "` + messageID + `"}`))
}