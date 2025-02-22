package router

import (
	"chat-service/handler"
	"chat-service/middleware"
	authpb "chat-service/proto/auth-service/proto"
	"chat-service/storage" // Импортируем вашу реализацию хранилища
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes устанавливает маршруты для чатов
func SetupRoutes(storage storage.Storage, authClient authpb.AuthServiceClient) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/chats", handler.CreateChatHandler(storage)).Methods("POST")
	router.HandleFunc("/api/chats", handler.GetUserChatsHandler(storage)).Methods("GET")
	router.Handle("/api/chats/{chatID}", middleware.AuthMiddleware(authClient, handler.UpdateChatHandler(storage))).Methods("PUT")
	router.HandleFunc("/api/chats/{chatID}", handler.DeleteChatHandler(storage)).Methods("DELETE")

	router.HandleFunc("/api/chats/{chatID}", func(w http.ResponseWriter, r *http.Request) {handler.GetChatByIDHandler(w, r, storage)}).Methods("GET")
	router.HandleFunc("/api/chats/{chatID}/avatar", func(w http.ResponseWriter, r *http.Request) {handler.SetChatAvatarHandler(w, r, storage)}).Methods("PUT")
	router.HandleFunc("/api/chats/{chatID}/participants", func(w http.ResponseWriter, r *http.Request) {handler.AddParticipantHandler(w, r, storage)}).Methods("POST")
	router.HandleFunc("/api/chats/{chatID}/participants", func(w http.ResponseWriter, r *http.Request) {handler.RemoveParticipantHandler(w, r, storage)}).Methods("DELETE")
	router.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {handler.SendMessageHandler(w, r, storage)}).Methods("POST")
	router.HandleFunc("/api/messages/{messageID}", func(w http.ResponseWriter, r *http.Request) {handler.EditMessageHandler(w, r, storage)}).Methods("PUT")
	router.HandleFunc("/api/messages/{messageID}", func(w http.ResponseWriter, r *http.Request) {handler.DeleteMessageHandler(w, r, storage)}).Methods("DELETE")
	router.HandleFunc("/api/chats/{chatID}/history", func(w http.ResponseWriter, r *http.Request) {handler.GetChatHistoryHandler(w, r, storage)}).Methods("GET")
	router.HandleFunc("/api/chats/{chatID}/participants", func(w http.ResponseWriter, r *http.Request) {handler.GetChatParticipantsHandler(w, r, storage)}).Methods("GET")
	router.HandleFunc("/api/chats/{chatID}/leave", func(w http.ResponseWriter, r *http.Request) {handler.LeaveChatHandler(w, r, storage)}).Methods("DELETE")
	router.HandleFunc("/api/messages/upload", func(w http.ResponseWriter, r *http.Request) {handler.UploadFileHandler(w, r, storage)}).Methods("POST")
	router.HandleFunc("/api/messages/{messageID}/reactions", func(w http.ResponseWriter, r *http.Request) {handler.AddReactionHandler(w, r, storage)}).Methods("POST")
	router.HandleFunc("/api/messages/{messageID}/reactions", func(w http.ResponseWriter, r *http.Request) {handler.RemoveReactionHandler(w, r, storage)}).Methods("DELETE")
	router.HandleFunc("/api/messages/{messageID}/status", func(w http.ResponseWriter, r *http.Request) {handler.UpdateMessageStatusHandler(w, r, storage)}).Methods("POST")

	return router
}
