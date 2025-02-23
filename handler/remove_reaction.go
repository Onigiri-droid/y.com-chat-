package handler

import (
	"encoding/json"
	"net/http"

	"chat-service/middleware"
	"chat-service/storage"

	"github.com/gorilla/mux"
)

// RemoveReactionHandler обрабатывает запросы на удаление реакции с сообщения.
func RemoveReactionHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        vars := mux.Vars(r)
        messageID := vars["messageID"]

        var req struct {
            Reaction string `json:"reaction"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Неверный формат данных", http.StatusBadRequest)
            return
        }

        userID, ok := ctx.Value(middleware.UserIDKey).(int32) // Получаем userID из контекста
        if !ok {
            http.Error(w, "Не удалось получить userID", http.StatusUnauthorized)
            return
        }

        if err := store.RemoveReaction(ctx, messageID, req.Reaction, userID); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "success",
            "reaction": req.Reaction,
        })
    }
}