package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// RemoveReactionRequest представляет данные запроса для удаления реакции
type RemoveReactionRequest struct {
	Reaction string `json:"reaction"` // Реакция (например, "👍", "❤️", "😂")
}

// RemoveReactionHandler обрабатывает удаление реакции с сообщения
func RemoveReactionHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodDelete {
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
	var req RemoveReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что reaction передан
	if req.Reaction == "" {
		http.Error(w, "Не указана реакция", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для удаления реакции
	err := storage.RemoveReaction(r.Context(), messageID, req.Reaction)
	if err != nil {
		http.Error(w, "Не удалось удалить реакцию: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Реакция удалена"}`))
}