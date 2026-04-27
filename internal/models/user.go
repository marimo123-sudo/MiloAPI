package models

import "time"

type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Password      string    `json:"-" db:"password"`
	ActiveGoalID  string    `json:"active_goal_id" db:"active_goal_id"`
	ActiveThemeID string    `json:"active_theme_id" db:"active_theme_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

func (u *User) InsertColumns() []string {
	return []string{"email", "password"}
}

func (u *User) InsertValues() []interface{} {
	return []interface{}{u.Email, u.Password}
}
