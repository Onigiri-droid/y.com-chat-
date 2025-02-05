package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// GetChatHistoryHandler обрабатывает запрос на получение истории сообщений чата
func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
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

	// Вызываем метод хранилища для получения истории сообщений
	messages, err := storage.GetMessages(r.Context(), chatID)
	if err != nil {
		http.Error(w, "Не удалось получить историю сообщений: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сериализуем список сообщений в JSON
	response, err := json.Marshal(messages)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}