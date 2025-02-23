package handler

import (
	"chat-service/middleware"
	"chat-service/storage"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadFileHandler обрабатывает запросы на загрузку файла.
func UploadFileHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Парсим multipart/form-data
        err := r.ParseMultipartForm(10 << 20) // Ограничение размера файла до 10 МБ
        if err != nil {
            log.Printf("Ошибка при парсинге формы: %v", err)
            http.Error(w, "Невозможно обработать запрос", http.StatusBadRequest)
            return
        }

        // Извлекаем файл из формы
        file, handler, err := r.FormFile("file")
        if err != nil {
            log.Printf("Ошибка при извлечении файла: %v", err)
            http.Error(w, "Не удалось получить файл", http.StatusBadRequest)
            return
        }
        defer file.Close()

        // Проверяем, что файл был передан
        if handler == nil {
            http.Error(w, "Файл не найден", http.StatusBadRequest)
            return
        }

        // Получаем chatID и senderID из контекста
        chatID := r.FormValue("chat_id")
        if chatID == "" {
            http.Error(w, "Не указан chatID", http.StatusBadRequest)
            return
        }

        userID, ok := r.Context().Value(middleware.UserIDKey).(int32)
        if !ok {
            log.Printf("Не удалось извлечь userID из контекста")
            http.Error(w, "Не удалось извлечь userID", http.StatusInternalServerError)
            return
        }

        // Создаем уникальное имя для файла
        ext := filepath.Ext(handler.Filename)
        fileName := fmt.Sprintf("%s%s", generateUniqueFileName(), ext)
        filePath := filepath.Join("uploads", fileName)

        // Создаем директорию для загрузки файлов, если она не существует
        if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
            log.Printf("Ошибка создания директории для загрузки файлов: %v", err)
            http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
            return
        }

        // Сохраняем файл на диск
        dst, err := os.Create(filePath)
        if err != nil {
            log.Printf("Ошибка создания файла: %v", err)
            http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        _, err = io.Copy(dst, file)
        if err != nil {
            log.Printf("Ошибка записи файла: %v", err)
            http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
            return
        }

        // Генерируем URL файла
        fileURL := fmt.Sprintf("/uploads/%s", fileName)

        // Создаем сообщение с типом "file"
        messageID, err := store.SaveMessage(ctx, chatID, userID, fileURL, "file")
        if err != nil {
            log.Printf("Ошибка сохранения сообщения: %v", err)
            http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{
            "message_id": messageID,
            "file_url":   fileURL,
            "status":     "success",
        })
    }
}

// Helper function to generate a unique file name
func generateUniqueFileName() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}