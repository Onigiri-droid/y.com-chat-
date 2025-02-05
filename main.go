package main

import (
	"chat-service/middleware"
	"chat-service/router"
	"chat-service/storage"
	chatpb "chat-service/proto/chat-service/proto"
	authpb "chat-service/proto/auth-service/proto"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Структура ChatService
type ChatService struct {
	chatpb.UnimplementedChatServiceServer
	MongoStorage *storage.MongoStorage
}

type contextKey string

const userIDKey contextKey = "user_id"

func tokenAuthInterceptor(authClient authpb.AuthServiceClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "метаданные не найдены")
		}

		authHeader, exists := md["authorization"]
		if !exists || len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "токен отсутствует")
		}

		// Проверяем токен через AuthService
		resp, err := authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: authHeader[0]})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "ошибка проверки токена")
		}

		if !resp.Valid {
			return nil, status.Error(codes.Unauthenticated, "токен недействителен")
		}

		// Добавляем user_id в контекст
		ctx = context.WithValue(ctx, userIDKey, resp.UserId)

		return handler(ctx, req)
	}
}

// Функция для подключения к gRPC-сервису с повторными попытками
func connectWithRetry(ctx context.Context, target string, retries int, delay time.Duration) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			conn, err = grpc.DialContext(
				ctx,
				target,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultCallOptions(
					grpc.MaxCallRecvMsgSize(100*1024*1024), // 100 MB
					grpc.MaxCallSendMsgSize(100*1024*1024), // 100 MB
				),
			)
			if err == nil {
				return conn, nil
			}
			log.Printf("Не удалось подключиться к %s: попытка %d, ошибка: %v", target, i+1, err)
			time.Sleep(delay)
		}
	}
	return nil, err
}

func main() {
	// Читаем настройки из переменных окружения
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongo-db:27017/?ssl=false"
	}
	dbName := "chat_service"
	apiServiceAddr := os.Getenv("API_SERVICE_ADDR")
	if apiServiceAddr == "" {
		apiServiceAddr = "api-service:50051" // gRPC должен идти на 50051
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081" // Убираем двоеточие
	} else if httpPort[0] == ':' {
		httpPort = httpPort[1:] // Убираем двоеточие, если есть
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	} else if grpcPort[0] == ':' {
		grpcPort = grpcPort[1:]
	}

	// Подключение к MongoDB
	mongoStorage, err := storage.NewMongoStorage(mongoURI, dbName)
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := mongoStorage.Close(context.Background()); err != nil {
			log.Printf("Ошибка при закрытии подключения к MongoDB: %v", err)
		}
	}()

	// Проверяем доступность MongoDB
	if err := mongoStorage.Ping(context.Background()); err != nil {
		log.Fatalf("MongoDB недоступна: %v", err)
	}

	// Подключение к AuthService (Api-service)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := connectWithRetry(ctx, apiServiceAddr, 10, 5*time.Second)
	if err != nil {
		log.Fatalf("Не удалось подключиться к Api-service: %v", err)
	}
	defer conn.Close()

	authClient := authpb.NewAuthServiceClient(conn)

	// Запуск gRPC-сервера
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(tokenAuthInterceptor(authClient)),
	)

	chatService := &ChatService{MongoStorage: mongoStorage}
	chatpb.RegisterChatServiceServer(grpcServer, chatService)

	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Ошибка запуска gRPC-сервера: %v", err)
	}

	go func() {
		log.Printf("gRPC-сервер запущен на %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
		}
	}()

	// Настройка HTTP-сервера
	mux := router.SetupRoutes(mongoStorage, authClient)
	handlerWithMiddleware := middleware.AuthMiddleware(authClient, mux)

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handlerWithMiddleware,
	}

	// Запуск HTTP-сервера
	go func() {
		log.Printf("HTTP-сервер запущен на %s", httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка работы HTTP-сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	// Завершаем работу серверов
	log.Println("Завершаем работу серверов...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Ошибка при завершении HTTP-сервера: %v", err)
	}

	grpcServer.GracefulStop()

	log.Println("Серверы успешно завершили работу")
}