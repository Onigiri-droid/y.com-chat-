package router

import (
	"chat-service/handler"
	"chat-service/middleware"
	authpb "chat-service/proto/auth-service/proto"
	"chat-service/storage" // Импортируем вашу реализацию хранилища

	"github.com/gorilla/mux"
)

// SetupRoutes устанавливает маршруты для чатов
func SetupRoutes(storage storage.Storage, authClient authpb.AuthServiceClient) *mux.Router {
	router := mux.NewRouter()
	// Создание нового чата
	router.HandleFunc("/api/chats", handler.CreateChatHandler(storage)).Methods("POST")
	// Получение списка чатов пользователя
	router.HandleFunc("/api/chats", handler.GetUserChatsHandler(storage)).Methods("GET")
	// Обновление информации о чате (требуется аутентификация)
	router.Handle("/api/chats/{chatID}", middleware.AuthMiddleware(authClient, handler.UpdateChatHandler(storage))).Methods("PUT")
	// Удаление чата
	router.HandleFunc("/api/chats/{chatID}", handler.DeleteChatHandler(storage)).Methods("DELETE")
	// Получение информации о чате по его ID
	router.HandleFunc("/api/chats/{chatID}", handler.GetChatByIDHandler(storage)).Methods("GET")
	// Установка или обновление аватара чата
	router.HandleFunc("/api/chats/{chatID}/avatar", handler.SetChatAvatarHandler(storage)).Methods("PUT")
	// Добавление участника в чат
	router.HandleFunc("/api/chats/{chatID}/participants", handler.AddParticipantHandler(storage)).Methods("POST")
	// Удаление участника из чата
	router.HandleFunc("/api/chats/{chatID}/participants", handler.RemoveParticipantHandler(storage)).Methods("DELETE")
	// Отправка сообщения в чат
	router.Handle("/api/messages", middleware.AuthMiddleware(authClient, handler.SendMessageHandler(storage))).Methods("POST")
	// Получение истории сообщений в чате
	router.Handle("/api/chats/{chatID}/history", handler.GetChatHistoryHandler(storage)).Methods("GET")
	// Редактирование сообщения по его ID
	router.HandleFunc("/api/messages/{messageID}", handler.EditMessageHandler(storage)).Methods("PUT")
	// Удаление сообщения по его ID
	router.HandleFunc("/api/messages/{messageID}", handler.DeleteMessageHandler(storage)).Methods("DELETE")
	// Получение списка участников чата
	router.HandleFunc("/api/chats/{chatID}/participants", handler.GetChatParticipantsHandler(storage)).Methods("GET")
	// Выход пользователя из чата
	router.HandleFunc("/api/chats/{chatID}/leave", handler.LeaveChatHandler(storage)).Methods("DELETE")
	// Загрузка файла в сообщение
	router.HandleFunc("/api/messages/upload", handler.UploadFileHandler(storage)).Methods("POST")
	// Добавление реакции на сообщение
	router.HandleFunc("/api/messages/{messageID}/reactions", handler.AddReactionHandler(storage)).Methods("POST")
	// Удаление реакции с сообщения
	router.HandleFunc("/api/messages/{messageID}/reactions", handler.RemoveReactionHandler(storage)).Methods("DELETE")
	// Обновление статуса сообщения (например, прочитано/доставлено)
	router.HandleFunc("/api/messages/{messageID}/status", handler.UpdateMessageStatusHandler(storage)).Methods("POST")
	return router
}