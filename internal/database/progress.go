package database

import (
	"api/internal/goalsdata"
	"api/internal/initialize" // добавлен импорт глобальной переменной
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type NextPhraseInfo struct {
	ID          int    `json:"id"`
	Phrase      string `json:"phrase"`
	Translation string `json:"translation"`
	GoalID      int    `json:"goal_id"`
	ThemeID     int    `json:"theme_id"`
}

// CompletedPhraseInfo содержит информацию о выученной фразе
type CompletedPhraseInfo struct {
	PhraseID    int       `json:"phrase_id"`
	Phrase      string    `json:"phrase"`
	Translation string    `json:"translation"`
	GoalID      int       `json:"goal_id"`
	ThemeID     int       `json:"theme_id"`
	CompletedAt time.Time `json:"completed_at"` // дата завершения
}

// GetAllCompletedPhrases возвращает все выученные пользователем фразы (is_completed = true).
func GetAllCompletedPhrases(userID string) ([]CompletedPhraseInfo, error) {
	// 1. Запрашиваем все завершённые записи прогресса для пользователя
	data := initialize.Data
	query := `SELECT phrase_id, goal_id, theme_id, created_at 
	          FROM progress
	          WHERE user_id = $1 AND is_completed = true
	          ORDER BY created_at ASC`

	rows, err := initialize.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query completed phrases: %w", err)
	}
	defer rows.Close()

	var results []CompletedPhraseInfo
	for rows.Next() {
		var phraseID, goalID, themeID int
		var completedAt time.Time
		if err := rows.Scan(&phraseID, &goalID, &themeID, &completedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 2. Получаем текст и перевод фразы из data
		goal, err := data.GetGoalByID(goalID)
		if err != nil {
			return nil, fmt.Errorf("goal %d not found for phrase %d: %w", goalID, phraseID, err)
		}
		theme, err := goal.GetThemeByID(themeID)
		if err != nil {
			return nil, fmt.Errorf("theme %d not found in goal %d for phrase %d: %w", themeID, goalID, phraseID, err)
		}
		phrase, err := theme.GetPhraseByID(phraseID)
		if err != nil {
			return nil, fmt.Errorf("phrase %d not found in theme %d: %w", phraseID, themeID, err)
		}

		results = append(results, CompletedPhraseInfo{
			PhraseID:    phraseID,
			Phrase:      phrase.Phrase,
			Translation: phrase.Translation,
			GoalID:      goalID,
			ThemeID:     themeID,
			CompletedAt: completedAt,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return results, nil
}

// MarkPhrasesAsCompleted отмечает переданные phrase_id как выполненные (is_completed = true).
// Параметры:
//   - ctx: контекст (например, r.Context())
//   - userID: идентификатор пользователя (UUID)
//   - goalID: идентификатор цели
//   - themeID: идентификатор темы
//   - phraseIDs: список идентификаторов фраз (целые числа)
func MarkPhrasesAsCompleted(ctx context.Context, userID string, goalID, themeID int, phraseIDs []int) error {
	if len(phraseIDs) == 0 {
		// ничего не делаем, если список пуст
		return nil
	}

	query := `
		UPDATE progress
		SET is_completed = true, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
		  AND goal_id = $2
		  AND theme_id = $3
		  AND phrase_id = ANY($4::int[])
	`
	_, err := initialize.DB.ExecContext(ctx, query, userID, goalID, themeID, pq.Array(phraseIDs))
	if err != nil {
		return err
	}
	return nil
}

// GetRemainingPhrasesInCurrentTheme возвращает количество фраз, оставшихся в текущей теме пользователя.
func GetRemainingPhrasesInCurrentTheme(userID string) (int, error) {
	data := initialize.Data
	goalID, themeID, phraseID, err := GetActivePhrase(userID)
	if err != nil {
		return 0, err
	}
	remaining := data.RemainingPhrasesInTheme(goalID, themeID, phraseID)
	if remaining < 0 {
		return 0, fmt.Errorf("текущая фраза не найдена в данных")
	}
	return remaining, nil
}

// GetNextPhrases возвращает до 10 следующих фраз после текущей активной фразы пользователя.
func GetNextPhrases(userID string) ([]NextPhraseInfo, error) {
	data := initialize.Data
	curGoal, curTheme, curPhrase, err := GetActivePhrase(userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить активную фразу: %w", err)
	}

	var result []NextPhraseInfo
	nextGoal, nextTheme, nextPhrase, found := goalsdata.GetNextPhrase(data, curGoal, curTheme, curPhrase)
	if !found {
		return result, nil
	}

	for i := 0; i < 10 && found; i++ {
		goal, err := data.GetGoalByID(nextGoal)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения цели %d: %w", nextGoal, err)
		}
		theme, err := goal.GetThemeByID(nextTheme)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения темы %d: %w", nextTheme, err)
		}
		phrase, err := theme.GetPhraseByID(nextPhrase)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения фразы %d: %w", nextPhrase, err)
		}

		result = append(result, NextPhraseInfo{
			ID:          phrase.ID,
			Phrase:      phrase.Phrase,
			Translation: phrase.Translation,
			GoalID:      nextGoal,
			ThemeID:     nextTheme,
		})

		nextGoal, nextTheme, nextPhrase, found = goalsdata.GetNextPhrase(data, nextGoal, nextTheme, nextPhrase)
	}
	return result, nil
}

// StartUserProgress создаёт первую активную фразу для пользователя (начало первой темы первой цели).
func StartUserProgress(userID uuid.UUID) error {
	data := initialize.Data
	if len(data.Goals) == 0 {
		return fmt.Errorf("no goals in data")
	}
	firstGoal := data.Goals[0]
	if len(firstGoal.Themes) == 0 {
		return fmt.Errorf("no themes in first goal")
	}
	firstTheme := firstGoal.Themes[0]
	if len(firstTheme.Phrases) == 0 {
		return fmt.Errorf("no phrases in first theme")
	}
	firstPhrase := firstTheme.Phrases[0]

	var exists bool
	err := initialize.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM progress WHERE user_id = $1 AND is_active = true)`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already has active progress")
	}

	_, err = initialize.DB.Exec(`INSERT INTO progress (user_id, goal_id, theme_id, phrase_id, is_completed, is_active)
		VALUES ($1, $2, $3, $4, false, true)`,
		userID, firstGoal.ID, firstTheme.ID, firstPhrase.ID)
	return err
}

// GetActivePhrase возвращает (goal_id, theme_id, phrase_id) для активной фразы пользователя.
func GetActivePhrase(userID string) (goalID, themeID, phraseID int, err error) {
	query := `SELECT goal_id, theme_id, phrase_id FROM progress 
		WHERE user_id = $1 AND is_active = true LIMIT 1`
	row := initialize.DB.QueryRow(query, userID)
	err = row.Scan(&goalID, &themeID, &phraseID)
	if err == sql.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("no active phrase for user %v", userID)
	}
	return
}

// CompleteCurrentPhrase обрабатывает завершение текущей активной фразы.
// Возвращает:
//   - status: "phrase_completed", "theme_completed", "goal_completed"
//   - newGoalID, newThemeID, newPhraseID (0 если статус goal_completed)
//   - error
func CompleteCurrentPhrase(userID uuid.UUID) (status string, newGoalID, newThemeID, newPhraseID int, err error) {
	data := initialize.Data
	tx, err := initialize.DB.Begin()
	if err != nil {
		return "", 0, 0, 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var curGoalID, curThemeID, curPhraseID int
	query := `SELECT goal_id, theme_id, phrase_id FROM progress 
		WHERE user_id = $1 AND is_active = true FOR UPDATE`
	row := tx.QueryRow(query, userID)
	err = row.Scan(&curGoalID, &curThemeID, &curPhraseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, 0, 0, fmt.Errorf("no active phrase for user %v", userID)
		}
		return "", 0, 0, 0, err
	}

	_, err = tx.Exec(`UPDATE progress SET is_completed = true, is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND phrase_id = $2 AND is_active = true`, userID, curPhraseID)
	if err != nil {
		return "", 0, 0, 0, err
	}

	nextGoalID, nextThemeID, nextPhraseID, found := goalsdata.GetNextPhrase(data, curGoalID, curThemeID, curPhraseID)
	if !found {
		return "goal_completed", 0, 0, 0, nil
	}

	_, err = tx.Exec(`INSERT INTO progress (user_id, goal_id, theme_id, phrase_id, is_completed, is_active)
		VALUES ($1, $2, $3, $4, false, true)`,
		userID, nextGoalID, nextThemeID, nextPhraseID)
	if err != nil {
		return "", 0, 0, 0, err
	}

	if curThemeID != nextThemeID {
		status = "theme_completed"
	} else {
		status = "phrase_completed"
	}
	return status, nextGoalID, nextThemeID, nextPhraseID, nil
}
