package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	authpb "chat-service/proto/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthMiddleware проверяет токен через AuthService
func AuthMiddleware(authClient authpb.AuthServiceClient, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Отсутствует токен аутентификации", http.StatusUnauthorized)
			return
		}

		// Проверяем формат токена (Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Некорректный формат токена", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// Логируем токен для отладки
		log.Printf("Получен токен: %s", token)

		// Вызываем ValidateToken через gRPC
		resp, err := authClient.ValidateToken(r.Context(), &authpb.ValidateTokenRequest{Token: token})
		if err != nil {
			// Логируем ошибку
			log.Printf("Ошибка ValidateToken: %v", err)

			// Обрабатываем gRPC-статусы
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.Unauthenticated:
					http.Error(w, "Невалидный токен", http.StatusUnauthorized)
				case codes.DeadlineExceeded, codes.Unavailable:
					http.Error(w, "Сервис аутентификации недоступен", http.StatusServiceUnavailable)
				default:
					http.Error(w, "Ошибка проверки токена", http.StatusInternalServerError)
				}
			} else {
				http.Error(w, "Неизвестная ошибка", http.StatusInternalServerError)
			}
			return
		}

		// Проверяем, валиден ли токен
		if !resp.Valid {
			log.Printf("Токен невалиден: %s", token)
			http.Error(w, "Невалидный токен", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст с использованием собственного типа ключа
		ctx := context.WithValue(r.Context(), UserIDKey, resp.UserId)
        next.ServeHTTP(w, r.WithContext(ctx))
	})
}