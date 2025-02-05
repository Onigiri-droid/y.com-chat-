package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client      *mongo.Client
	chatColl    *mongo.Collection
	messageColl *mongo.Collection
}

const (
	chatsCollection    = "chats"
	messagesCollection = "messages"
)

func NewMongoStorage(uri string, dbName string) (*MongoStorage, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	db := client.Database(dbName)
	return &MongoStorage{
		client:      client,
		chatColl:    db.Collection(chatsCollection),
		messageColl: db.Collection(messagesCollection),
	}, nil
}

func (m *MongoStorage) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *MongoStorage) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

func (m *MongoStorage) CreateChat(ctx context.Context, name string, memberIDs []int32, isGroup bool, description string) (string, error) {
	chat := bson.M{
		"name":        name,
		"member_ids":  memberIDs, // Участники чата, включая создателя
		"is_group":    isGroup,
		"description": description,
		"created_at":  time.Now(),
	}

	res, err := m.chatColl.InsertOne(ctx, chat)
	if err != nil {
		log.Printf("Ошибка создания чата: %v", err)
		return "", err
	}

	objectID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("не удалось преобразовать InsertedID в ObjectID")
	}

	return objectID.Hex(), nil
}

func (m *MongoStorage) UpdateChatInfo(ctx context.Context, chatID string, name string, description string) error {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("некорректный chatID")
	}

	// Проверяем, что передано хотя бы одно поле для обновления
	if name == "" && description == "" {
		return errors.New("не переданы данные для обновления")
	}

	// Создаём объект для обновления
	update := bson.M{}
	if name != "" {
		update["name"] = name
	}
	if description != "" {
		update["description"] = description
	}

	// Обновляем чат (только групповые чаты)
	res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID, "is_group": true}, bson.M{"$set": update})
	if err != nil {
		log.Printf("Ошибка обновления чата: %v", err)
		return errors.New("ошибка обновления чата")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("чат не найден или не является групповым")
	}

	return nil
}

func (m *MongoStorage) SetChatAvatar(ctx context.Context, chatID string, avatar string) error {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("некорректный chatID")
	}

	// Проверяем, что avatar не пустой
	if avatar == "" {
		return errors.New("не указан аватар")
	}

	// Обновляем аватар чата
	update := bson.M{"$set": bson.M{"avatar": avatar}}

	// Выполняем обновление
	res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка обновления аватара чата: %v", err)
		return errors.New("ошибка обновления аватара чата")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("чат не найден")
	}

	return nil
}

func (m *MongoStorage) AddParticipant(ctx context.Context, chatID string, userID int32) error {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("некорректный chatID")
	}

	// Проверяем, что userID не равен 0
	if userID == 0 {
		return errors.New("некорректный userID")
	}

	// Обновляем список участников чата
	update := bson.M{"$addToSet": bson.M{"member_ids": userID}}

	// Выполняем обновление (только для групповых чатов)
	res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID, "is_group": true}, update)
	if err != nil {
		log.Printf("Ошибка добавления участника: %v", err)
		return errors.New("ошибка добавления участника")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("чат не найден или не является групповым")
	}

	return nil
}

func (m *MongoStorage) RemoveParticipant(ctx context.Context, chatID string, userID int32) error {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("некорректный chatID")
	}

	// Проверяем, что userID не равен 0
	if userID == 0 {
		return errors.New("некорректный userID")
	}

	// Удаляем участника из списка участников чата
	update := bson.M{"$pull": bson.M{"member_ids": userID}}

	// Выполняем обновление (только для групповых чатов)
	res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID, "is_group": true}, update)
	if err != nil {
		log.Printf("Ошибка удаления участника: %v", err)
		return errors.New("ошибка удаления участника")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("чат не найден или не является групповым")
	}

	return nil
}

func (m *MongoStorage) SaveMessage(ctx context.Context, chatID string, senderID int32, content string, messageType string) (string, error) {
	chatObjectID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Printf("Некорректный идентификатор чата: %v", err)
		return "", errors.New("некорректный идентификатор чата")
	}

	message := bson.M{
		"chat_id":    chatObjectID,
		"sender_id":  senderID,
		"content":    content,
		"type":       messageType, // Тип сообщения (например, "file", "image", "video", "audio")
		"created_at": time.Now(),
	}

	res, err := m.messageColl.InsertOne(ctx, message)
	if err != nil {
		log.Printf("Ошибка сохранения сообщения: %v", err)
		return "", err
	}

	objectID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("не удалось преобразовать InsertedID в ObjectID")
	}

	return objectID.Hex(), nil
}

func (m *MongoStorage) EditMessage(ctx context.Context, messageID string, content string) error {
	// Преобразуем messageID в ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return errors.New("некорректный messageID")
	}

	// Проверяем, что content не пустой
	if content == "" {
		return errors.New("не указано содержимое сообщения")
	}

	// Обновляем содержимое сообщения
	update := bson.M{"$set": bson.M{"content": content}}

	// Выполняем обновление
	res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка редактирования сообщения: %v", err)
		return errors.New("ошибка редактирования сообщения")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("сообщение не найдено")
	}

	return nil
}

func (m *MongoStorage) DeleteMessage(ctx context.Context, messageID string) error {
	// Преобразуем messageID в ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return errors.New("некорректный messageID")
	}

	// Удаляем сообщение
	res, err := m.messageColl.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
		return errors.New("ошибка удаления сообщения")
	}

	// Проверяем, был ли удалён хотя бы один документ
	if res.DeletedCount == 0 {
		return errors.New("сообщение не найдено")
	}

	return nil
}

func (m *MongoStorage) GetUserChats(ctx context.Context, userID int32) ([]*Chat, error) {
	// Ищем чаты, в которых состоит пользователь
	filter := bson.M{"member_ids": userID}

	cursor, err := m.chatColl.Find(ctx, filter)
	if err != nil {
		log.Printf("Ошибка получения списка чатов: %v", err)
		return nil, errors.New("ошибка получения списка чатов")
	}
	defer cursor.Close(ctx)

	var chats []*Chat
	for cursor.Next(ctx) {
		var chat Chat
		if err := cursor.Decode(&chat); err != nil {
			log.Printf("Ошибка декодирования чата: %v", err)
			return nil, err
		}
		chats = append(chats, &chat)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Ошибка работы курсора: %v", err)
		return nil, err
	}

	return chats, nil
}

func (m *MongoStorage) GetMessages(ctx context.Context, chatID string) ([]*Message, error) {
	// Преобразуем chatID в ObjectID
	chatObjectID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Printf("Некорректный идентификатор чата: %v", err)
		return nil, errors.New("некорректный идентификатор чата")
	}

	// Ищем сообщения для указанного чата
	cursor, err := m.messageColl.Find(ctx, bson.M{"chat_id": chatObjectID})
	if err != nil {
		log.Printf("Ошибка получения сообщений: %v", err)
		return nil, errors.New("ошибка получения сообщений")
	}
	defer cursor.Close(ctx)

	var messages []*Message
	for cursor.Next(ctx) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			log.Printf("Ошибка декодирования сообщения: %v", err)
			return nil, err
		}
		messages = append(messages, &msg)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Ошибка работы курсора: %v", err)
		return nil, err
	}

	return messages, nil
}

func (m *MongoStorage) GetChatParticipants(ctx context.Context, chatID string) ([]int32, error) {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, errors.New("некорректный chatID")
	}

	// Ищем чат по ID
	var chat Chat
	err = m.chatColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("чат не найден")
		}
		log.Printf("Ошибка получения чата: %v", err)
		return nil, errors.New("ошибка получения чата")
	}

	// Возвращаем список участников
	return chat.MemberIDs, nil
}

func (m *MongoStorage) LeaveChat(ctx context.Context, chatID string, userID int32) error {
	// Преобразуем chatID в ObjectID
	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("некорректный chatID")
	}

	// Проверяем, что userID не равен 0
	if userID == 0 {
		return errors.New("некорректный userID")
	}

	// Удаляем пользователя из списка участников чата
	update := bson.M{"$pull": bson.M{"member_ids": userID}}

	// Выполняем обновление
	res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка выхода из чата: %v", err)
		return errors.New("ошибка выхода из чата")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("чат не найден")
	}

	return nil
}

func (m *MongoStorage) AddReaction(ctx context.Context, messageID string, reaction string) error {
	// Преобразуем messageID в ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return errors.New("некорректный messageID")
	}

	// Проверяем, что reaction не пустая
	if reaction == "" {
		return errors.New("не указана реакция")
	}

	// Добавляем реакцию к сообщению
	update := bson.M{"$push": bson.M{"reactions": reaction}}

	// Выполняем обновление
	res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка добавления реакции: %v", err)
		return errors.New("ошибка добавления реакции")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("сообщение не найдено")
	}

	return nil
}

func (m *MongoStorage) RemoveReaction(ctx context.Context, messageID string, reaction string) error {
	// Преобразуем messageID в ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return errors.New("некорректный messageID")
	}

	// Проверяем, что reaction не пустая
	if reaction == "" {
		return errors.New("не указана реакция")
	}

	// Удаляем реакцию из сообщения
	update := bson.M{"$pull": bson.M{"reactions": reaction}}

	// Выполняем обновление
	res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка удаления реакции: %v", err)
		return errors.New("ошибка удаления реакции")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("сообщение не найдено")
	}

	return nil
}

func (m *MongoStorage) UpdateMessageStatus(ctx context.Context, messageID string, status string) error {
	// Преобразуем messageID в ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return errors.New("некорректный messageID")
	}

	// Проверяем, что status не пустой
	if status == "" {
		return errors.New("не указан статус")
	}

	// Обновляем статус сообщения
	update := bson.M{"$set": bson.M{"status": status}}

	// Выполняем обновление
	res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		log.Printf("Ошибка обновления статуса сообщения: %v", err)
		return errors.New("ошибка обновления статуса сообщения")
	}

	// Проверяем, был ли обновлён хотя бы один документ
	if res.MatchedCount == 0 {
		return errors.New("сообщение не найдено")
	}

	return nil
}

func (m *MongoStorage) GetChatByID(ctx context.Context, chatID string) (*Chat, error) {
	// Преобразуем chatID в ObjectID
	chatObjectID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Printf("Некорректный идентификатор чата: %v", err)
		return nil, errors.New("некорректный идентификатор чата")
	}

	// Ищем чат по ID
	var chat Chat
	err = m.chatColl.FindOne(ctx, bson.M{"_id": chatObjectID}).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("чат не найден")
		}
		log.Printf("Ошибка получения чата: %v", err)
		return nil, err
	}

	// Возвращаем найденный чат
	return &chat, nil
}

func (m *MongoStorage) DeleteChat(ctx context.Context, chatID string) error {
	// Преобразуем chatID в ObjectID
	chatObjectID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Printf("Некорректный идентификатор чата: %v", err)
		return errors.New("некорректный идентификатор чата")
	}

	// Удаляем все сообщения, связанные с этим чатом
	_, err = m.messageColl.DeleteMany(ctx, bson.M{"chat_id": chatObjectID})
	if err != nil {
		log.Printf("Ошибка удаления сообщений чата: %v", err)
		return errors.New("ошибка удаления сообщений чата")
	}

	// Удаляем сам чат
	res, err := m.chatColl.DeleteOne(ctx, bson.M{"_id": chatObjectID})
	if err != nil {
		log.Printf("Ошибка удаления чата: %v", err)
		return err
	}

	// Проверяем, был ли удалён хотя бы один документ
	if res.DeletedCount == 0 {
		return errors.New("чат не найден")
	}

	return nil
}

// Типы данных
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ChatID    primitive.ObjectID `bson:"chat_id"`
	SenderID  int32              `bson:"sender_id"`
	Content   string             `bson:"content"`   // Содержимое сообщения
	Type      string             `bson:"type"`      // Тип сообщения (например, "text", "file", "image", "video", "audio")
	Reactions []string           `bson:"reactions"` // Список реакций
	Status    string             `bson:"status"`    // Статус сообщения (например, "delivered", "read")
	CreatedAt time.Time          `bson:"created_at"`
}

type Chat struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description,omitempty"`
	Avatar      string             `bson:"avatar,omitempty"`
	MemberIDs   []int32            `bson:"member_ids"`
	IsGroup     bool               `bson:"is_group"`
	CreatedAt   time.Time          `bson:"created_at"`
}

// Storage - интерфейс для работы с хранилищем.
type Storage interface {
	CreateChat(ctx context.Context, name string, memberIDs []int32, isGroup bool, description string) (string, error)
	UpdateChatInfo(ctx context.Context, chatID string, name string, description string) error
	SetChatAvatar(ctx context.Context, chatID string, avatar string) error
	AddParticipant(ctx context.Context, chatID string, userID int32) error
	RemoveParticipant(ctx context.Context, chatID string, userID int32) error
	SaveMessage(ctx context.Context, chatID string, senderID int32, content string, messageType string) (string, error)
	EditMessage(ctx context.Context, messageID string, content string) error
	DeleteMessage(ctx context.Context, messageID string) error
	GetUserChats(ctx context.Context, userID int32) ([]*Chat, error)
	GetMessages(ctx context.Context, chatID string) ([]*Message, error)
	GetChatParticipants(ctx context.Context, chatID string) ([]int32, error)
	LeaveChat(ctx context.Context, chatID string, userID int32) error
	AddReaction(ctx context.Context, messageID string, reaction string) error
	RemoveReaction(ctx context.Context, messageID string, reaction string) error
	UpdateMessageStatus(ctx context.Context, messageID string, status string) error

	GetChatByID(ctx context.Context, chatID string) (*Chat, error)
	DeleteChat(ctx context.Context, chatID string) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}
