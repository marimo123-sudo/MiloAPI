package handler

import (
	"api/internal/database"
	"api/internal/models"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)
	list_of_users, err := database.GetAllUsers()
	if err != nil {
		SendJSONError(
			w,
			map[string]any{
				"status":  "error",
				"message": "Couldn't get list of users from Database",
			},
			http.StatusInternalServerError,
		)
		return
	}
	sendJSONSuccess(w, map[string]interface{}{
		"status":        "ok",
		"status_code":   200,
		"list_of_users": list_of_users,
	})

}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)
	params := mux.Vars(r)
	user_id := params["id"]
	if _, err := uuid.Parse(user_id); err != nil {
		SendJSONError(w, map[string]interface{}{"status": err, "message": "Invalid user_id in request keys"}, http.StatusBadRequest)
		return
	}
	user_info, err := database.GetUserInfo(user_id)
	if err != nil {
		SendJSONError(
			w,
			map[string]any{
				"status":  "error",
				"message": "Couldn't get User information from Database",
			},
			http.StatusInternalServerError,
		)
		return
	}
	list_of_users_words, err := database.GetAllCompletedPhrases(user_id)
	sendJSONSuccess(
		w,
		map[string]interface{}{
			"status":      "ok",
			"status_code": 200,
			"user":        user_info,
			"user_words":  list_of_users_words,
		},
	)
}

func Register(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		}, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Email and password are required",
		}, http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		}, http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hashed),
	}

	userID, createdAt, err := database.InsertUser(user)
	if err != nil {
		// Предполагаем, что ошибка может быть из-за дубликата email
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "User with this email already exists",
		}, http.StatusConflict)
		return
	}

	sendJSONSuccess(w, map[string]interface{}{
		"status":      "ok",
		"status_code": 201,
		"user_id":     userID,
		"created_at":  createdAt,
		"message":     "User registered successfully",
	})
}

// Login аутентифицирует пользователя и возвращает его ID (без токена)
func Login(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	user, err := database.GetUserByEmail(req.Email)
	if err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid email or password",
		}, http.StatusUnauthorized)
		return
	}

	// Сравниваем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid email or password",
		}, http.StatusUnauthorized)
		return
	}

	sendJSONSuccess(w, map[string]interface{}{
		"status":      "ok",
		"status_code": 200,
		"user_id":     user.ID,
		"email":       user.Email,
		"message":     "Login successful",
	})
}

func UsersHandler(r *mux.Router) {
	r.HandleFunc("/users", GetAllUsers).Methods("GET")
	r.HandleFunc("/user/{id}", GetUserInfo).Methods("GET")
	r.HandleFunc("/register", Register).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
}
