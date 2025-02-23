package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"chat-service/middleware"
	"chat-service/storage"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetChatHistoryHandler handles requests to retrieve chat history.
func GetChatHistoryHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Проверяем метод запроса
        if r.Method != http.MethodGet {
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        // Извлекаем chatID из URL
        chatID := extractChatIDFromURL(r.URL.Path)
        if !isValidChatID(chatID) {
            http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            return
        }

        // Извлекаем userID из контекста
        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
            return
        }

        // Получаем информацию о чате
        chat, err := store.GetChatByID(ctx, chatID)
        if err != nil {
            handleChatError(w, err)
            return
        }

        // Проверяем права пользователя на просмотр чата
        if !isUserAllowedToViewChat(chat, userID) {
            http.Error(w, "У вас нет прав на просмотр этого чата", http.StatusForbidden)
            return
        }

        // Получаем параметры пагинации
        limit, skip, err := parsePaginationParams(r.URL.Query())
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Получаем историю сообщений
        var messages []*storage.Message
        if limit > 0 || skip > 0 {
            messages, err = store.GetMessagesWithPagination(ctx, chatID, limit, skip)
        } else {
            messages, err = store.GetMessages(ctx, chatID)
        }
        if err != nil {
            handleMessageError(w, err)
            return
        }

        // Возвращаем успешный ответ с историей сообщений
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(messages); err != nil {
            log.Printf("Ошибка при кодировании JSON: %v", err)
            http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
        }
    }
}

// Helper Functions
func extractChatIDFromURL(path string) string {
    return strings.TrimSuffix(strings.TrimPrefix(path, "/api/chats/"), "/history")
}

func isValidChatID(chatID string) bool {
    return primitive.IsValidObjectID(chatID)
}

func parsePaginationParams(queryParams map[string][]string) (int64, int64, error) {
    var limit, skip int64

    if values, ok := queryParams["limit"]; ok && len(values) > 0 {
        var err error
        limit, err = strconv.ParseInt(values[0], 10, 64)
        if err != nil || limit < 0 {
            return 0, 0, errors.New("некорректный параметр 'limit'")
        }
    }

    if values, ok := queryParams["skip"]; ok && len(values) > 0 {
        var err error
        skip, err = strconv.ParseInt(values[0], 10, 64)
        if err != nil || skip < 0 {
            return 0, 0, errors.New("некорректный параметр 'skip'")
        }
    }

    return limit, skip, nil
}

func isUserAllowedToViewChat(chat *storage.Chat, userID int32) bool {
    for _, memberID := range chat.MemberIDs {
        if memberID == userID {
            return true
        }
    }
    return false
}

func handleChatError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, storage.ErrInvalidChatID):
        http.Error(w, "Некорректный chatID", http.StatusBadRequest)
    case errors.Is(err, storage.ErrChatNotFound):
        http.Error(w, "Чат не найден", http.StatusNotFound)
    default:
        log.Printf("Ошибка получения чата: %v", err)
        http.Error(w, "Не удалось получить информацию о чате", http.StatusInternalServerError)
    }
}

func handleMessageError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, storage.ErrInvalidChatID):
        http.Error(w, "Некорректный chatID", http.StatusBadRequest)
    case errors.Is(err, storage.ErrMessageNotFound):
        http.Error(w, "Сообщения не найдены", http.StatusNotFound)
    default:
        log.Printf("Ошибка получения сообщений: %v", err)
        http.Error(w, "Неизвестная ошибка при получении сообщений", http.StatusInternalServerError)
    }
}
