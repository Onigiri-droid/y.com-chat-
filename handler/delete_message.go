package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// DeleteMessageHandler обрабатывает запросы на удаление сообщения
func DeleteMessageHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Получение ID сообщения из URL
        vars := mux.Vars(r)
        messageID := vars["messageID"]

        // Получение userID из контекста
        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID", http.StatusInternalServerError)
            return
        }

        // Вызов метода для удаления сообщения
        err := store.DeleteMessage(ctx, messageID, userID)
        if err != nil {
            switch {
            case errors.Is(err, storage.ErrInvalidMessageID):
                http.Error(w, "Некорректный messageID", http.StatusBadRequest)
            case errors.Is(err, storage.ErrMessageNotFound):
                http.Error(w, "Сообщение не найдено", http.StatusNotFound)
            default:
                log.Printf("Ошибка при удалении сообщения: %v", err)
                http.Error(w, "Ошибка при удалении сообщения", http.StatusInternalServerError)
            }
            return
        }

        // Успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "success",
            "message": "Сообщение удалено",
        })
    }
}