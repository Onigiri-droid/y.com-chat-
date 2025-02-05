package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// RemoveParticipantRequest представляет данные запроса для удаления участника
type RemoveParticipantRequest struct {
	UserID int32 `json:"userID"` // ID пользователя, которого нужно удалить
}

// RemoveParticipantHandler обрабатывает удаление участника из группового чата
func RemoveParticipantHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodDelete {
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
	var req RemoveParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что userID передан
	if req.UserID == 0 {
		http.Error(w, "Не указан userID", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для удаления участника
	err := storage.RemoveParticipant(r.Context(), chatID, req.UserID)
	if err != nil {
		http.Error(w, "Не удалось удалить участника: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Участник удалён"}`))
}