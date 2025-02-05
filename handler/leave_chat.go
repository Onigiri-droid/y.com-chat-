package handler

import (
	"chat-service/storage"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// LeaveChatHandler обрабатывает запрос на выход пользователя из чата
func LeaveChatHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
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

	// Извлекаем userID из запроса (например, из токена или query-параметра)
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Не указан userID", http.StatusBadRequest)
		return
	}

	// Преобразуем userID в int32
	var userIDInt int32
	_, err := fmt.Sscanf(userID, "%d", &userIDInt)
	if err != nil {
		http.Error(w, "Некорректный userID", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для выхода пользователя из чата
	err = storage.LeaveChat(r.Context(), chatID, userIDInt)
	if err != nil {
		http.Error(w, "Не удалось выйти из чата: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Пользователь вышел из чата"}`))
}