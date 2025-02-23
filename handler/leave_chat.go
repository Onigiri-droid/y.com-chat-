package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// LeaveChatHandler обрабатывает запросы на выход пользователя из чата.
func LeaveChatHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Извлекаем chatID из URL
        vars := mux.Vars(r)
        chatID := vars["chatID"]

        // Получаем userID из контекста
        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID", http.StatusInternalServerError)
            return
        }

        // Вызываем метод для выхода из чата
        err := store.LeaveChat(ctx, chatID, userID)
        if err != nil {
            switch {
            case errors.Is(err, storage.ErrInvalidChatID):
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case errors.Is(err, storage.ErrChatNotFound):
                http.Error(w, "Чат не найден", http.StatusNotFound)
            default:
                log.Printf("Ошибка выхода из чата: %v", err)
                http.Error(w, "Ошибка выхода из чата", http.StatusInternalServerError)
            }
            return
        }

        // Возвращаем успешный ответ
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Вы успешно покинули чат"))
    }
}