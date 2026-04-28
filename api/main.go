package main

import (
	"api/internal/handler"
	"api/internal/initialize"
	"log"
	"net/http"
	"sync"
)

var (
	router http.Handler
	once   sync.Once
)

func init() {
	once.Do(func() {
		if err := initialize.EnvInitialization(); err != nil {
			log.Fatal(err)
		}
		router = handler.HandleAllFuncs()
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

