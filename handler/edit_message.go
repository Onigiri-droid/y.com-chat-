package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// EditMessageRequest представляет данные запроса для редактирования сообщения
type EditMessageRequest struct {
	Content string `json:"content"` // Новое содержимое сообщения
}

// EditMessageHandler обрабатывает редактирование сообщения
func EditMessageHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPut {
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
	var req EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что content передан
	if req.Content == "" {
		http.Error(w, "Не указано содержимое сообщения", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для редактирования сообщения
	err := storage.EditMessage(r.Context(), messageID, req.Content)
	if err != nil {
		http.Error(w, "Не удалось отредактировать сообщение: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Сообщение отредактировано"}`))
}