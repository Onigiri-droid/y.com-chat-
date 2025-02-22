package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"log"
	"net/http"
)

// isUserAllowedToDeleteChat проверяет, имеет ли пользователь права на удаление чата
func isUserAllowedToDeleteChat(chat *storage.Chat, userID int32) bool {
    // Если это групповой чат, проверяем, является ли пользователь создателем
    if chat.IsGroup {
        return chat.CreatorID == userID
    }

    // Для личных чатов проверяем, является ли пользователь участником
    for _, memberID := range chat.MemberIDs {
        if memberID == userID {
            return true
        }
    }

    return false
}

func DeleteChatHandler(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Проверяем метод запроса
        if r.Method != http.MethodDelete {
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        // Извлекаем chatID из URL-параметров
        chatID := r.URL.Path[len("/api/chats/"):]
        if chatID == "" {
            http.Error(w, "Отсутствует обязательный параметр chatID", http.StatusBadRequest)
            return
        }

        // Извлекаем userID из контекста (добавленного AuthMiddleware)
        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID из токена", http.StatusInternalServerError)
            return
        }

        // Получаем информацию о чате
        chat, err := storage.GetChatByID(r.Context(), chatID)
        if err != nil {
            switch err.Error() {
            case "некорректный идентификатор чата":
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case "чат не найден":
                http.Error(w, "Чат не найден", http.StatusNotFound)
            default:
                log.Printf("Ошибка получения чата: %v", err)
                http.Error(w, "Не удалось получить информацию о чате", http.StatusInternalServerError)
            }
            return
        }

        // Проверяем права пользователя на удаление чата
        if !isUserAllowedToDeleteChat(chat, userID) {
            http.Error(w, "У вас нет прав на удаление этого чата", http.StatusForbidden)
            return
        }

        // Вызываем метод хранилища для удаления чата
        err = storage.DeleteChat(r.Context(), chatID)
        if err != nil {
            switch err.Error() {
            case "некорректный идентификатор чата":
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case "чат не найден":
                http.Error(w, "Чат не найден", http.StatusNotFound)
            default:
                log.Printf("Ошибка при удалении чата: %v", err)
                http.Error(w, "Не удалось удалить чат", http.StatusInternalServerError)
            }
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(map[string]string{"message": "Chat deleted successfully"}); err != nil {
            http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
        }
    }
}