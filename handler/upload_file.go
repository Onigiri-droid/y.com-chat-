package handler

import (
	"encoding/json"
	"net/http"
	"chat-service/storage"
)

// UploadFileRequest представляет данные запроса для отправки файла
type UploadFileRequest struct {
	ChatID string `json:"chatID"` // ID чата, в который отправляется файл
	File   string `json:"file"`   // Base64-кодированный файл
	Type   string `json:"type"`   // Тип файла (например, "file", "image", "video", "audio")
}

// UploadFileHandler обрабатывает отправку файлов и медиа в чат
func UploadFileHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем тело запроса
	var req UploadFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Проверяем, что все обязательные поля переданы
	if req.ChatID == "" || req.File == "" || req.Type == "" {
		http.Error(w, "Не указаны chatID, file или type", http.StatusBadRequest)
		return
	}

	// Вызываем метод хранилища для сохранения файла
	messageID, err := storage.SaveMessage(r.Context(), req.ChatID, 0, req.File, req.Type) // senderID пока не используется
	if err != nil {
		http.Error(w, "Не удалось отправить файл: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ с ID сообщения
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"messageID": "` + messageID + `"}`))
}