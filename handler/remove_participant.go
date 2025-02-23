package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// RemoveParticipantHandler обрабатывает удаление участника из группового чата
func RemoveParticipantHandler(storage storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Проверяем метод запроса
        if r.Method != http.MethodDelete {
            http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
            return
        }

        // Извлекаем chatID из URL-параметров
        chatID := strings.TrimPrefix(r.URL.Path, "/api/chats/")
        chatID = strings.TrimSuffix(chatID[:len(chatID)-13], "/") // Удаляем "/participants"
        if chatID == "" || len(chatID) != 24 {
            http.Error(w, "Некорректный chatID", http.StatusBadRequest)
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

        // Проверяем права пользователя на управление участниками чата
        if !isUserAllowedToManageParticipants(chat, userID) {
            http.Error(w, "У вас нет прав на управление участниками этого чата", http.StatusForbidden)
            return
        }

        // Читаем тело запроса
        var req struct {
            UserID int32 `json:"user_id"` // ID удаляемого пользователя
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Некорректный запрос", http.StatusBadRequest)
            return
        }

        // Проверяем, что передано значение для userID
        if req.UserID == 0 {
            http.Error(w, "Не указан ID пользователя для удаления", http.StatusBadRequest)
            return
        }

        // Вызываем метод хранилища для удаления участника
        err = storage.RemoveParticipant(r.Context(), chatID, req.UserID)
        if err != nil {
            switch err.Error() {
            case "некорректный chatID":
                http.Error(w, "Некорректный chatID", http.StatusBadRequest)
            case "некорректный userID":
                http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
            case "чат не найден или не является групповым":
                http.Error(w, "Чат не найден или не является групповым", http.StatusNotFound)
            default:
                log.Printf("Ошибка при удалении участника: %v", err)
                http.Error(w, "Не удалось удалить участника", http.StatusInternalServerError)
            }
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(map[string]string{"message": "Participant removed successfully"}); err != nil {
            http.Error(w, "Не удалось отправить ответ", http.StatusInternalServerError)
        }
    }
}