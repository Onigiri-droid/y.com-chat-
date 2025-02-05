package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// UpdateMessageStatusRequest представляет данные запроса для обновления статуса сообщения
type UpdateMessageStatusRequest struct {
	Status string `json:"status"` // Статус сообщения (например, "delivered", "read")
}

// UpdateMessageStatusHandler обрабатывает обновление статуса сообщения
func UpdateMessageStatusHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем messageID из URL
	vars := mux.Vars(r)
	messageID, exists := vars["messageID"]
	if !exists {
		http.Error(w, "Не указан messageID", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req UpdateMessageStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что status передан
	if req.Status == "" {
		http.Error(w, "Не указан статус", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для обновления статуса сообщения
	err := storage.UpdateMessageStatus(r.Context(), messageID, req.Status)
	if err != nil {
		http.Error(w, "Не удалось обновить статус сообщения: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Статус сообщения обновлён"}`))
}