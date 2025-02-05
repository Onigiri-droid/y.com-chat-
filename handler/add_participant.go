package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// AddParticipantRequest представляет данные запроса для добавления участника
type AddParticipantRequest struct {
	UserID int32 `json:"userID"` // ID пользователя, которого нужно добавить
}

// AddParticipantHandler обрабатывает добавление участника в групповой чат
func AddParticipantHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
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
	var req AddParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что userID передан
	if req.UserID == 0 {
		http.Error(w, "Не указан userID", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для добавления участника
	err := storage.AddParticipant(r.Context(), chatID, req.UserID)
	if err != nil {
		http.Error(w, "Не удалось добавить участника: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Участник добавлен"}`))
}