package handler

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"chat-service/storage"
)

// SetChatAvatarRequest представляет данные запроса для установки аватара чата
type SetChatAvatarRequest struct {
	Avatar string `json:"avatar"` // Base64-кодированное изображение
}

// SetChatAvatarHandler обрабатывает установку аватара чата
func SetChatAvatarHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPut {
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
	var req SetChatAvatarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что поле avatar передано
	if req.Avatar == "" {
		http.Error(w, "Не указан аватар", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для установки аватара
	err := storage.SetChatAvatar(r.Context(), chatID, req.Avatar)
	if err != nil {
		http.Error(w, "Не удалось установить аватар: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Аватар чата обновлён"}`))
}