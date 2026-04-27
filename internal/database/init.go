package database

import (
	"api/internal/initialize"
	"log"
	"reflect"
	"strings"
)

func GetStructFields(model interface{}) map[string]bool {
	fields := make(map[string]bool)
	t := reflect.TypeOf(model)

	// Если передан указатель, получаем значение
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Проходим по всем полям структуры
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Получаем тег db
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			fields[dbTag] = true
		}
		// Также добавляем имя поля в нижнем регистре
		fields[strings.ToLower(field.Name)] = true
	}

	return fields
}

func CreateTables() error {

	queries := []string{
		// Расширение UUID (если не установлено)
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		// Таблица пользователей (минимальная для аутентификации)
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			conversation_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Единая таблица прогресса
		`CREATE TABLE IF NOT EXISTS progress (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL,
			goal_id INTEGER NOT NULL,
			theme_id INTEGER NOT NULL,
			phrase_id INTEGER NOT NULL,
			is_completed BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Уникальность: у одного пользователя не может быть дублей (user_id + phrase_id)
		`ALTER TABLE progress ADD CONSTRAINT unique_user_phrase UNIQUE (user_id, phrase_id)`,
	}

	for _, q := range queries {
		if _, err := initialize.DB.Exec(q); err != nil {
			log.Printf("❌ Ошибка выполнения запроса: %v\nЗапрос: %s", err, q)
			return err
		}
	}

	// Таблицы user_goals, user_themes, user_phrases_progress больше не нужны
	log.Println("✅ Таблицы users и progress успешно созданы")
	return nil
}
