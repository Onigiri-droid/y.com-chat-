package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// GetChatByIDHandler обрабатывает запрос на получение информации о чате
func GetChatByIDHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
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

	// Вызываем метод хранилища для получения информации о чате
	chat, err := storage.GetChatByID(r.Context(), chatID)
	if err != nil {
		http.Error(w, "Не удалось получить информацию о чате: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сериализуем информацию о чате в JSON
	response, err := json.Marshal(chat)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}