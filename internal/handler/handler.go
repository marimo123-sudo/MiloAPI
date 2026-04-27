package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func SendJSONError(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SetJSONAnswer(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func sendJSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func send_hello(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)
	sendJSONSuccess(w, map[string]any{
		"status":      "ok",
		"status_code": 200,
	})
}

func HandleAllFuncs() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.HandleFunc("/", send_hello).Methods("GET")
	r.Use(loggingMiddleware)

	UsersHandler(r)
	log.Println("Server starting on :8443")
	log.Fatal(http.ListenAndServe(":8443", r))
}
