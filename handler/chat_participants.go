package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"chat-service/storage"
)

type contextKey string

const userIDKey contextKey = "userID"

// GetChatParticipantsHandler обрабатывает запрос на получение списка участников чата
func GetChatParticipantsHandler(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем userID из контекста
	userID, ok := r.Context().Value(userIDKey).(int32)
	if !ok {
		log.Printf("Не удалось извлечь userID из контекста")
		http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
		return
	}

	// Логируем userID для отладки
	log.Printf("Получен запрос на список чатов для userID: %d", userID)

	// Вызываем метод хранилища для получения списка чатов
	chats, err := storage.GetUserChats(r.Context(), userID)
	if err != nil {
		log.Printf("Ошибка при получении списка чатов: %v", err)
		http.Error(w, "Не удалось получить список чатов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем количество найденных чатов
	log.Printf("Найдено %d чатов для userID: %d", len(chats), userID)

	// Сериализуем список чатов в JSON
	response, err := json.Marshal(chats)
	if err != nil {
		log.Printf("Ошибка сериализации данных: %v", err)
		http.Error(w, "Ошибка сериализации данных", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}