package database

import (
	"api/internal/ai"
	"api/internal/initialize"
	"api/internal/models"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

func GetUserInfo(user_id string) (models.User, error) {
	query := "SELECT * FROM users WHERE id = $1"
	var user models.User
	err := initialize.DB.Get(&user, query, user_id)
	if err != nil {
		log.Printf("Ошибка при получении пользователя (%v): %e", user_id, err)
	}
	return user, err
}

func GetUserConversation(ctx context.Context, c *ai.Client, userID string) (string, error) {
	var conversationID string
	err := initialize.DB.Get(&conversationID, "SELECT conversation_id FROM users WHERE id = $1", userID)
	if err != nil {
		log.Printf("Ошибка получения conversation_id: %v", err)
		return "", err
	}
	if conversationID != "" {
		return conversationID, nil
	}
	newConvID, err := c.NewConversation(ctx)
	if err != nil {
		log.Printf("Ошибка создания диалога: %v", err)
		return "", err
	}

	// Сохраняем в БД
	_, err = initialize.DB.Exec("UPDATE users SET conversation_id = $1 WHERE id = $2", newConvID, userID)
	if err != nil {
		log.Printf("Ошибка сохранения conversation_id: %v", err)
		return "", err
	}

	return newConvID, nil
}

// DeleteUserConversation удаляет conversation_id у пользователя.
// Если нужно также удалить сам диалог в OpenAI, вызовите соответствующий метод API,
// но обычно достаточно очистить поле в БД.
func DeleteUserConversation(userID string) error {
	_, err := initialize.DB.Exec("UPDATE users SET conversation_id = '' WHERE id = $1", userID)
	if err != nil {
		log.Printf("Ошибка удаления conversation_id: %v", err)
		return err
	}
	log.Printf("Conversation_id удалён у пользователя %s", userID)
	return nil

}

func GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := initialize.DB.Select(&users, "SELECT * FROM users")
	if err != nil {
		fmt.Println("Не удалось получить информацию о всех пользователях", err)
		return nil, err
	}
	return users, nil
}

func InsertUser(user models.User) (string, time.Time, error) {
	columns := user.InsertColumns()
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf(
		"INSERT INTO users (%s) VALUES (%s) RETURNING id, created_at",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	values := user.InsertValues()
	err := initialize.DB.QueryRowx(query, values...).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Не удалось Добавить Пользователя %v", user.ID)
	}
	return user.ID, user.CreatedAt, err
}

func GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE email = $1"
	err := initialize.DB.Get(&user, query, email)
	if err != nil {
		fmt.Println("Не удалось Получить пользователя по Email")
	}
	return user, err
}

func UpdateUserField(id string, column string, value interface{}) error {
	var user models.User
	valid_columns := GetStructFields(user)
	delete(valid_columns, "id")
	delete(valid_columns, "created_at")
	if !valid_columns[column] {
		return fmt.Errorf("колонка '%s' не существует в таблице tg_bots", column)
	}
	query := fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2", column)

	_, err := initialize.DB.Exec(query, value, id)
	if err != nil {
		fmt.Println("Не получилось обновить пользователя:", err)
	}
	return err
}

func IsPasswordCorrect(user_id string, password string) (bool, error) {
	query := "SELECT password FROM users WHERE id = $1"
	var real_password string
	err := initialize.DB.Get(&real_password, query, user_id)

	if err != nil {
		log.Printf("Не удалось получить пароль пользователя (%v): %e", user_id, err)
		return false, err
	}
	if real_password == password {
		return true, nil
	}
	return false, nil

}

func CreateNewUserGoal(user_id string, goal_id int) error {
	query := "INSERT INTO user_goals (user_id, goal_id) VALUES ($1, $2)"
	_, err := initialize.DB.Exec(query, user_id, goal_id)
	if err != nil {
		log.Println("Не удалось добавить Цель Пользователю:", err)
	}
	return err
}
