package middleware

// Определяем собственный тип для ключа контекста
type contextKey string

// Константа для ключа
const (
    UserIDKey contextKey = "userID"
    RoleKey   contextKey = "role"
    // Другие ключи...
)