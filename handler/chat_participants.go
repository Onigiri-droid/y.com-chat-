package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"chat-service/storage"

	"github.com/gorilla/mux"
)

// GetChatParticipantsHandler обрабатывает запрос на получение списка участников чата
func GetChatParticipantsHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Извлекаем chatID из URL
        vars := mux.Vars(r)
        chatID := vars["chatID"]

        // Получаем список участников чата
        participants, err := store.GetChatParticipants(ctx, chatID)
        if err != nil {
            switch {
            case errors.Is(err, storage.ErrInvalidChatID):
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case errors.Is(err, storage.ErrChatNotFound):
                http.Error(w, "Чат не найден", http.StatusNotFound)
            default:
                log.Printf("Ошибка получения участников чата: %v", err)
                http.Error(w, "Ошибка получения участников чата", http.StatusInternalServerError)
            }
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(participants)
    }
}