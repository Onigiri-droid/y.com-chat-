package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"chat-service/middleware"
	"chat-service/storage"
)

// CreateChatRequest представляет данные запроса для создания чата
type CreateChatRequest struct {
	Name         string   `json:"name"`         // Название чата
	Participants []string `json:"participants"` // Участники чата (ID в строковом виде)
	IsGroup      bool     `json:"is_group"`     // Групповой ли чат
	Description  string   `json:"description,omitempty"` // Описание чата (необязательное)
}

// CreateChatResponse представляет ответ при создании чата
type CreateChatResponse struct {
	ChatID string `json:"chat_id"` // Идентификатор созданного чата
}

// CreateChatHandler обрабатывает запросы на создание чатов
func CreateChatHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		// Извлекаем userID из контекста (добавленного AuthMiddleware)
		userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
		if !ok {
			log.Printf("Не удалось извлечь userID из контекста")
			http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
			return
		}
		// Используем userID
		log.Printf("UserID: %d", userID)

		// Читаем тело запроса
		var req CreateChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}

		// Проверяем обязательные поля
		if req.Name == "" || len(req.Participants) == 0 {
			http.Error(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
			return
		}

		// Конвертация Participants из []string в []int32
		var memberIDs []int32
		for _, p := range req.Participants {
			id, err := strconv.Atoi(p)
			if err != nil {
				http.Error(w, "Некорректный формат participant ID", http.StatusBadRequest)
				return
			}
			memberIDs = append(memberIDs, int32(id))
		}

		// Добавляем userID создателя в список участников
		memberIDs = append(memberIDs, userID)

		// Для личных чатов (is_group: false) проверяем, что участников ровно два
		if !req.IsGroup && len(memberIDs) != 2 {
			http.Error(w, "Личный чат должен содержать ровно двух участников", http.StatusBadRequest)
			return
		}

		// Создаём чат в хранилище
		chatID, err := storage.CreateChat(r.Context(), req.Name, memberIDs, req.IsGroup, req.Description)
		if err != nil {
			http.Error(w, "Не удалось создать чат: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем успешный ответ
		resp := CreateChatResponse{ChatID: chatID}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
		}
	}
}