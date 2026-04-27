package handler

import (
	"api/internal/ai"
	"api/internal/database"
	"api/internal/initialize"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func GetAnswer(w http.ResponseWriter, r *http.Request) {
	SetJSONAnswer(w)
	ctx := r.Context()
	var Question struct {
		UserID string `json:"user_id"`
		Text   string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&Question); err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		}, http.StatusBadRequest)
		return
	}
	if Question.UserID == "" || Question.Text == "" {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "user_id and question are required",
		}, http.StatusBadRequest)
		return
	}
	var list_of_words []string

	if words, exists := initialize.UsersWords[Question.UserID]; exists {
		list_of_words = words
		// ключ существует, можно использовать list_of_words
	} else {
		// ключа нет – обработать ситуацию (например, создать пустой слайс)

		list_of_phrases, _ := database.GetNextPhrases(Question.UserID)
		list_of_words = []string{}
		for _, phrase := range list_of_phrases {
			final_string := fmt.Sprintf("%s - %s", phrase.Phrase, phrase.Translation)
			list_of_words = append(list_of_words, final_string)
		}
		initialize.UsersWords[Question.UserID] = list_of_words
	}

	AIClient := ai.NewClient()
	convID, err := database.GetUserConversation(ctx, AIClient, Question.UserID)
	answer, err := AIClient.GetAnswerFromAI(ctx, Question.Text, convID)
	if err != nil {
		SendJSONError(w, map[string]interface{}{
			"status":  "error",
			"message": "AI couldn't give answer",
		}, http.StatusBadRequest)
		return
	}
	end_of_dialog := false
	if strings.Contains(answer, "end_of_dialog") {
		list_of_phrases_ids := []int{}
		list_of_phrases, err1 := database.GetNextPhrases(Question.UserID)
		if err1 != nil {
			SendJSONError(w, map[string]interface{}{
				"status":  "error",
				"message": "Couldn't get users wordlist",
			}, http.StatusInternalServerError)
		}
		goal_id := list_of_phrases[0].GoalID
		theme_id := list_of_phrases[0].ThemeID
		for _, phrase := range list_of_phrases {
			list_of_phrases_ids = append(list_of_phrases_ids, phrase.ID)
		}
		end_of_dialog = true
		log.Println(list_of_phrases_ids)
		err := database.MarkPhrasesAsCompleted(ctx, Question.UserID, goal_id, theme_id, list_of_phrases_ids)
		if err != nil {
			SendJSONError(w, map[string]interface{}{
				"status":  "error",
				"message": "Couldn't mark new phrases",
			}, http.StatusInternalServerError)
			return
		}
		initialize.UsersWords[Question.UserID] = []string{}
	}
	sendJSONSuccess(w, map[string]interface{}{
		"status":        "ok",
		"status_code":   200,
		"answer":        answer,
		"end_of_dialog": end_of_dialog,
		"list_of_words": list_of_words,
	})

}

func HandleAIDialog(r *mux.Router) {
	r.HandleFunc("/get_answer", GetAnswer).Methods("GET")

}
