package handler

import (
	"chat-service/storage"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func UpdateMessageStatusHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Извлекаем messageID из URL
        vars := mux.Vars(r)
        messageID := vars["messageID"]
        log.Printf("Получен запрос на обновление статуса для messageID: %s", messageID)

        // Декодируем тело запроса
        var req struct {
            Status string `json:"status"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            log.Printf("Ошибка декодирования тела запроса: %v", err)
            http.Error(w, "Неверный формат данных", http.StatusBadRequest)
            return
        }

        log.Printf("Получен статус: %s", req.Status)

        // Обновляем статус сообщения
        if err := store.UpdateMessageStatus(ctx, messageID, req.Status); err != nil {
            log.Printf("Ошибка обновления статуса: %v", err)
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "success",
            "message": "Статус сообщения успешно обновлен",
        })
    }
}