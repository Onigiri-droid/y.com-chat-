package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// UpdateChatRequest представляет данные запроса для обновления чата
type UpdateChatRequest struct {
	Name        string `json:"name,omitempty"`        // Новое название чата (опционально)
	Description string `json:"description,omitempty"` // Новое описание чата (опционально)
}

// UpdateChatHandler обрабатывает обновление информации о групповом чате
func UpdateChatHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем chatID из URL
	vars := mux.Vars(r)
	chatID, exists := vars["chatID"]
	if !exists {
		http.Error(w, "Не указан chatID", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req UpdateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что хотя бы одно поле для обновления передано
	if req.Name == "" && req.Description == "" {
		http.Error(w, "Не указаны данные для обновления", http.StatusBadRequest)
		return
	}

	// Вызываем обновление в хранилище
	err := storage.UpdateChatInfo(r.Context(), chatID, req.Name, req.Description)
	if err != nil {
		http.Error(w, "Не удалось обновить чат: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Чат обновлён"}`))
}