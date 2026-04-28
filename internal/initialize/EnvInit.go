package initialize

import (
	"api/internal/config"
	"api/internal/goalsdata"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // правильный драйвер
	"github.com/jmoiron/sqlx"
)

var CFG *config.Config
var UsersWords map[string][]string
var DB *sqlx.DB
var Data *goalsdata.Data

func ConnectToDB() *sqlx.DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("БД не отвечает: ", err)
	}
	log.Println("✅ Подключено к базе данных Supabase")
	return db
}

func EnvInitialization() error {
	DB = ConnectToDB()
	// Если используете embed — без аргументов
	d, err := goalsdata.LoadData("internal/goalsdata/goals.json")
	if err != nil {
		return fmt.Errorf("загрузка данных: %w", err)
	}
	Data = &d
	CFG = config.Load()
	if CFG.OpenAIKey == "" {
		log.Fatal("OpenAIKey is required")
	}
	if CFG.Port == "" {
		CFG.Port = "8080"
	}
	return nil
}
