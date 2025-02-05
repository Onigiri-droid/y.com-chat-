package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// AddReactionRequest представляет данные запроса для добавления реакции
type AddReactionRequest struct {
	Reaction string `json:"reaction"` // Реакция (например, "👍", "❤️", "😂")
}

// AddReactionHandler обрабатывает добавление реакции к сообщению
func AddReactionHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
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
	var req AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что reaction передан
	if req.Reaction == "" {
		http.Error(w, "Не указана реакция", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для добавления реакции
	err := storage.AddReaction(r.Context(), messageID, req.Reaction)
	if err != nil {
		http.Error(w, "Не удалось добавить реакцию: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Реакция добавлена"}`))
}