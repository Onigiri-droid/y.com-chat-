package storage

import (
	"context"
	"errors"
	"log"
	"strings"
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

func (m *MongoStorage) CreateChat(ctx context.Context, name string, memberIDs []int32, isGroup bool, description string, creatorID int32) (string, error) {
    if creatorID == 0 {
        return "", errors.New("не указан создатель чата")
    }

    chat := bson.M{
        "name":        name,
        "member_ids":  memberIDs,
        "creator_id":  creatorID, // Сохраняем ID создателя чата
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

    // Обновляем чат (без проверки на is_group)
    res, err := m.chatColl.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
    if err != nil {
        log.Printf("Ошибка обновления чата: %v", err)
        return errors.New("ошибка обновления чата")
    }

    // Проверяем, был ли обновлён хотя бы один документ
    if res.MatchedCount == 0 {
        return errors.New("чат не найден")
    }

    return nil
}

func (m *MongoStorage) SetChatAvatar(ctx context.Context, chatID string, avatar string) error {
    // Преобразуем chatID в ObjectID
    objID, err := primitive.ObjectIDFromHex(chatID)
    if err != nil {
        log.Printf("Некорректный идентификатор чата: %v", err)
        return errors.New("некорректный идентификатор чата")
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

    // Проверяем, был ли обновлен хотя бы один документ
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
        "type":       messageType, // Например, "file"
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

func (m *MongoStorage) EditMessage(ctx context.Context, messageID string, userID int32, newContent string) error {
    // Преобразуем messageID в ObjectID
    objID, err := primitive.ObjectIDFromHex(messageID)
    if err != nil {
        return errors.New("некорректный messageID")
    }

    // Проверяем, что новое содержимое не пустое
    if strings.TrimSpace(newContent) == "" {
        return errors.New("содержимое сообщения не может быть пустым")
    }

    // Находим сообщение и проверяем автора
    var message struct {
        SenderID int32 `bson:"sender_id"`
    }
    err = m.messageColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&message)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return errors.New("сообщение не найдено")
        }
        log.Printf("Ошибка получения сообщения: %v", err)
        return errors.New("ошибка получения сообщения")
    }

    // Проверяем, что редактирует автор сообщения
    if message.SenderID != userID {
        return errors.New("вы не можете редактировать чужое сообщение")
    }

    // Обновляем содержимое сообщения и добавляем время редактирования
    update := bson.M{
        "$set": bson.M{
            "content":   newContent,
            "edited_at": time.Now(),
        },
    }

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

// DeleteMessage удаляет сообщение по его ID и ID пользователя.
func (m *MongoStorage) DeleteMessage(ctx context.Context, messageID string, userID int32) error {
    // Преобразование в ObjectID
    objID, err := primitive.ObjectIDFromHex(messageID)
    if err != nil {
        return ErrInvalidMessageID
    }

    // Условие удаления: сообщение должно принадлежать пользователю
    filter := bson.M{
        "_id":      objID,
        "sender_id": userID, // Обратите внимание на "sender_id" (нижнее подчеркивание)
    }

    // Выполнение удаления
    result, err := m.messageColl.DeleteOne(ctx, filter)
    if err != nil {
        return err
    }

    // Проверка, было ли сообщение удалено
    if result.DeletedCount == 0 {
        return ErrMessageNotFound
    }

    return nil
}

func (m *MongoStorage) GetUserChats(ctx context.Context, userID int32) ([]*Chat, error) {
    // Ищем чаты, где пользователь является участником или создателем
    filter := bson.M{
        "$or": []bson.M{
            {"member_ids": userID},
            {"creator_id": userID},
        },
    }

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

// GetMessages возвращает сообщения чата с возможностью пагинации.
func (m *MongoStorage) GetMessages(ctx context.Context, chatID string) ([]*Message, error) {
    chatObjectID, err := primitive.ObjectIDFromHex(chatID)
    if err != nil {
        log.Printf("Некорректный идентификатор чата: %v", err)
        return nil, errors.New("некорректный идентификатор чата")
    }

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

// GetMessagesWithPagination возвращает сообщения чата с параметрами пагинации.
func (m *MongoStorage) GetMessagesWithPagination(ctx context.Context, chatID string, limit int64, skip int64) ([]*Message, error) {
    chatObjectID, err := primitive.ObjectIDFromHex(chatID)
    if err != nil {
        log.Printf("Некорректный идентификатор чата: %v", err)
        return nil, errors.New("некорректный идентификатор чата")
    }

    findOptions := options.Find()
    if limit > 0 {
        findOptions.SetLimit(limit)
    }
    if skip > 0 {
        findOptions.SetSkip(skip)
    }

    cursor, err := m.messageColl.Find(ctx, bson.M{"chat_id": chatObjectID}, findOptions)
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
    objID, err := primitive.ObjectIDFromHex(chatID)
    if err != nil {
        return nil, ErrInvalidChatID
    }

    var chat Chat
    err = m.chatColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&chat)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrChatNotFound
        }
        return nil, err
    }

    return chat.MemberIDs, nil
}

func (m *MongoStorage) LeaveChat(ctx context.Context, chatID string, userID int32) error {
    // Преобразуем chatID в ObjectID
    objID, err := primitive.ObjectIDFromHex(chatID)
    if err != nil {
        return ErrInvalidChatID
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
        return ErrChatNotFound
    }

    return nil
}

func (m *MongoStorage) AddReaction(ctx context.Context, messageID string, reaction string, userID int32) error {
    objID, err := primitive.ObjectIDFromHex(messageID)
    if err != nil {
        return errors.New("некорректный messageID")
    }

    if reaction == "" {
        return errors.New("не указана реакция")
    }

    // Проверяем, что пользователь еще не добавил такую реакцию
    filter := bson.M{
        "_id": objID,
        "reactions": bson.M{
            "$elemMatch": bson.M{
                "reaction": reaction,
                "user_id":  userID,
            },
        },
    }

    var exists bool
    err = m.messageColl.FindOne(ctx, filter).Decode(&exists)
    if err == nil {
        return errors.New("реакция уже добавлена этим пользователем")
    }

    // Добавляем реакцию
    update := bson.M{"$push": bson.M{"reactions": bson.M{"reaction": reaction, "user_id": userID}}}
    res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        log.Printf("Ошибка добавления реакции: %v", err)
        return errors.New("ошибка добавления реакции")
    }

    if res.MatchedCount == 0 {
        return errors.New("сообщение не найдено")
    }

    return nil
}

func (m *MongoStorage) RemoveReaction(ctx context.Context, messageID string, reaction string, userID int32) error {
    objID, err := primitive.ObjectIDFromHex(messageID)
    if err != nil {
        return errors.New("некорректный messageID")
    }

    if reaction == "" {
        return errors.New("не указана реакция")
    }

    // Удаляем реакцию, если она была добавлена этим пользователем
    update := bson.M{"$pull": bson.M{"reactions": bson.M{"reaction": reaction, "user_id": userID}}}
    res, err := m.messageColl.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        log.Printf("Ошибка удаления реакции: %v", err)
        return errors.New("ошибка удаления реакции")
    }

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
    // Проверяем длину chatID
    if chatID == "" || len(chatID) != 24 {
        return nil, errors.New("некорректный идентификатор чата")
    }

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
        return nil, errors.New("ошибка получения чата")
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

var (
    ErrInvalidChatID   = errors.New("некорректный chatID")
    ErrChatNotFound    = errors.New("чат не найден")
    ErrInvalidMessageID = errors.New("некорректный messageID")
    ErrMessageNotFound = errors.New("сообщение не найдено")
    ErrForbidden = errors.New("чужое сообщение")
)

// Типы данных
type Reaction struct {
    Reaction string `bson:"reaction"`
    UserID   int32  `bson:"user_id"`
}

type Message struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    ChatID    primitive.ObjectID `bson:"chat_id"`
    SenderID  int32              `bson:"sender_id"`
    Content   string             `bson:"content"`
    Type      string             `bson:"type"`
    Reactions []Reaction         `bson:"reactions"`
    Status    string             `bson:"status"`
    CreatedAt time.Time          `bson:"created_at"`
}

type MessageReaction struct {
    MessageID primitive.ObjectID `bson:"message_id"` // ID сообщения
    Reactions []Reaction         `bson:"reactions"`  // Список реакций
}

type Chat struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Name        string             `bson:"name"`
    Description string             `bson:"description,omitempty"`
    Avatar      string             `bson:"avatar,omitempty"`
    CreatorID   int32              `bson:"creator_id"`
    MemberIDs   []int32            `bson:"member_ids"`
    IsGroup     bool               `bson:"is_group"`
    CreatedAt   time.Time          `bson:"created_at"`
}

// Storage - интерфейс для работы с хранилищем.
type Storage interface {
    CreateChat(ctx context.Context, name string, memberIDs []int32, isGroup bool, description string, creatorID int32) (string, error)
    UpdateChatInfo(ctx context.Context, chatID string, name string, description string) error
    SetChatAvatar(ctx context.Context, chatID string, avatar string) error
    AddParticipant(ctx context.Context, chatID string, userID int32) error
    RemoveParticipant(ctx context.Context, chatID string, userID int32) error
    SaveMessage(ctx context.Context, chatID string, senderID int32, content string, messageType string) (string, error)
    EditMessage(ctx context.Context, messageID string, userID int32, newContent string) error
    DeleteMessage(ctx context.Context, messageID string, userID int32) error
    GetUserChats(ctx context.Context, userID int32) ([]*Chat, error)
    GetMessages(ctx context.Context, chatID string) ([]*Message, error)
    GetMessagesWithPagination(ctx context.Context, chatID string, limit int64, skip int64) ([]*Message, error)
    GetChatParticipants(ctx context.Context, chatID string) ([]int32, error)
    LeaveChat(ctx context.Context, chatID string, userID int32) error
    AddReaction(ctx context.Context, messageID string, reaction string, userID int32) error
    RemoveReaction(ctx context.Context, messageID string, reaction string, userID int32) error
    UpdateMessageStatus(ctx context.Context, messageID string, status string) error
    GetChatByID(ctx context.Context, chatID string) (*Chat, error)
    DeleteChat(ctx context.Context, chatID string) error
    Close(ctx context.Context) error
    Ping(ctx context.Context) error
}