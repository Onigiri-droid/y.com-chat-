package handler

import (
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// DeleteChatHandler обрабатывает запрос на удаление чата
func DeleteChatHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
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

	// Вызываем метод хранилища для удаления чата
	err := storage.DeleteChat(r.Context(), chatID)
	if err != nil {
		http.Error(w, "Не удалось удалить чат: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Чат удалён"}`))
}