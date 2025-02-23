package handler

import (
	"chat-service/storage"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

// SetChatAvatarHandler обрабатывает запросы на установку или обновление аватара чата.
func SetChatAvatarHandler(store storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Извлекаем chatID из URL
        vars := mux.Vars(r)
        chatID := vars["chatID"]

        // Парсим multipart/form-data
        err := r.ParseMultipartForm(10 << 20) // Ограничение размера файла до 10 МБ
        if err != nil {
            log.Printf("Ошибка при парсинге формы: %v", err)
            http.Error(w, "Невозможно обработать запрос", http.StatusBadRequest)
            return
        }

        // Извлекаем файл из формы
        file, handler, err := r.FormFile("avatar") // Убедитесь, что ключ "avatar" совпадает с ключом в форме
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

        // Создаем уникальное имя для файла
        ext := filepath.Ext(handler.Filename)
        fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
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

        // Обновляем аватар чата в базе данных
        err = store.SetChatAvatar(ctx, chatID, fileURL)
        if err != nil {
            log.Printf("Ошибка обновления аватара чата: %v", err)
            http.Error(w, "Ошибка обновления аватара чата", http.StatusInternalServerError)
            return
        }

        // Возвращаем успешный ответ
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "avatar_url": fileURL,
            "status":     "success",
        })
    }
}