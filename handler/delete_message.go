package handler

import (
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// DeleteMessageHandler обрабатывает удаление сообщения
func DeleteMessageHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
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

	// Вызываем метод хранилища для удаления сообщения
	err := storage.DeleteMessage(r.Context(), messageID)
	if err != nil {
		http.Error(w, "Не удалось удалить сообщение: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Сообщение удалено"}`))
}