package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Подключение к MongoDB
func ConnectMongo(uri string, dbName string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uri)

	// Устанавливаем таймаут на подключение
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MongoDB: %v", err)
	}

	// Проверяем соединение
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к MongoDB: %v", err)
	}

	// Выполняем тестовый вызов "ping"
	err = client.Database(dbName).RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
	if err != nil {
		return nil, fmt.Errorf("MongoDB откликнулась, но команда ping вернула ошибку: %v", err)
	}

	log.Println("Подключение к MongoDB успешно установлено")
	return &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}

// Закрытие соединения с MongoDB
func (db *MongoDB) Close(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}
