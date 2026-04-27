package initialize

import (
	"api/internal/config"
	"api/internal/goalsdata"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var CFG *config.Config
var UsersWords map[string][]string
var DB *sqlx.DB
var Data *goalsdata.Data

func ConnectToDB() *sqlx.DB {
	connStr := "host=localhost port=5432 user=postgres dbname=server sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("БД не отвечает: ", err)
	}
	return db
}

func EnvInitialization() error {
	DB = ConnectToDB()
	d, err := goalsdata.LoadData("../../internal/goalsdata/goals.json")
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
