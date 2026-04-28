package main

import (
	"api/internal/handler"
	"api/internal/initialize"
	"log"
	"net/http"
)

func main() {
	if err := initialize.EnvInitialization(); err != nil {
		log.Fatal(err)
	}
	// Эндпоинт для получения временного токена (безопасная передача клиенту)
	handler.HandleAllFuncs()

	log.Printf("Server started at http://localhost:%s", initialize.CFG.Port)
	log.Fatal(http.ListenAndServe(":"+initialize.CFG.Port, nil))

}
