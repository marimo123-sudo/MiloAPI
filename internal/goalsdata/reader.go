package goalsdata

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Структуры, соответствующие JSON
type Phrase struct {
	ID          int    `json:"id"`
	Phrase      string `json:"phrase"`
	Translation string `json:"translation"`
}

type Theme struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Phrases []Phrase `json:"phrases"`
}

type Goal struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Themes []Theme `json:"themes"`
}

type Data struct {
	Goals []Goal `json:"goals"`
}

// LoadData загружает и парсит JSON файл
func LoadData(filename string) (Data, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Data{}, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return Data{}, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var data Data
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return Data{}, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}
	return data, nil
}

// GetGoals возвращает все цели
func (d *Data) GetGoals() []Goal {
	return d.Goals
}

// GetGoalByID находит цель по ID
func (d *Data) GetGoalByID(id int) (*Goal, error) {
	for i := range d.Goals {
		if d.Goals[i].ID == id {
			return &d.Goals[i], nil
		}
	}
	return nil, fmt.Errorf("цель с ID %d не найдена", id)
}

// GetThemeByID находит тему внутри цели по ID
func (g *Goal) GetThemeByID(themeID int) (*Theme, error) {
	for i := range g.Themes {
		if g.Themes[i].ID == themeID {
			return &g.Themes[i], nil
		}
	}
	return nil, fmt.Errorf("тема с ID %d не найдена в цели '%s'", themeID, g.Name)
}

// GetPhraseByID находит фразу внутри темы по ID
func (t *Theme) GetPhraseByID(phraseID int) (*Phrase, error) {
	for i := range t.Phrases {
		if t.Phrases[i].ID == phraseID {
			return &t.Phrases[i], nil
		}
	}
	return nil, fmt.Errorf("фраза с ID %d не найдена в теме '%s'", phraseID, t.Name)
}

// GetTranslation возвращает перевод фразы по трём ID
func (d *Data) GetTranslation(goalID, themeID, phraseID int) (string, error) {
	goal, err := d.GetGoalByID(goalID)
	if err != nil {
		return "", err
	}
	theme, err := goal.GetThemeByID(themeID)
	if err != nil {
		return "", err
	}
	phrase, err := theme.GetPhraseByID(phraseID)
	if err != nil {
		return "", err
	}
	return phrase.Translation, nil
}

// GetAllPhrases возвращает все фразы из всех тем (плоский список)
func (d *Data) GetAllPhrases() []Phrase {
	var all []Phrase
	for _, goal := range d.Goals {
		for _, theme := range goal.Themes {
			all = append(all, theme.Phrases...)
		}
	}
	return all
}

// NextPhrase возвращает ID следующей фразы в той же теме, где находится phraseID.
// Если phraseID не найден или это последняя фраза в теме — возвращает -1.
func (d *Data) NextPhrase(phraseID int) int {
	for _, goal := range d.Goals {
		for _, theme := range goal.Themes {
			for i, phrase := range theme.Phrases {
				if phrase.ID == phraseID {
					// если есть следующая фраза в этой теме
					if i+1 < len(theme.Phrases) {
						return theme.Phrases[i+1].ID
					}
					return -1
				}
			}
		}
	}
	return -2 // фраза не найдена
}

// GetNextPhrase ищет следующую фразу в структуре Data.
// Возвращает (goalID, themeID, phraseID, found). Если found == false – данных больше нет.
func GetNextPhrase(data *Data, curGoalID, curThemeID, curPhraseID int) (int, int, int, bool) {
	// Ищем текущую цель, тему и фразу
	for gIdx, goal := range data.Goals {
		if goal.ID == curGoalID {
			for tIdx, theme := range goal.Themes {
				if theme.ID == curThemeID {
					for pIdx, phrase := range theme.Phrases {
						if phrase.ID == curPhraseID {
							// Следующая фраза в той же теме
							if pIdx+1 < len(theme.Phrases) {
								next := theme.Phrases[pIdx+1]
								return goal.ID, theme.ID, next.ID, true
							}
							// Следующая тема в этой же цели
							for nextT := tIdx + 1; nextT < len(goal.Themes); nextT++ {
								if len(goal.Themes[nextT].Phrases) > 0 {
									nextTheme := goal.Themes[nextT]
									return goal.ID, nextTheme.ID, nextTheme.Phrases[0].ID, true
								}
							}
							// Следующая цель (после текущей)
							for nextG := gIdx + 1; nextG < len(data.Goals); nextG++ {
								nextGoal := data.Goals[nextG]
								for _, nextTheme := range nextGoal.Themes {
									if len(nextTheme.Phrases) > 0 {
										return nextGoal.ID, nextTheme.ID, nextTheme.Phrases[0].ID, true
									}
								}
							}
							return 0, 0, 0, false
						}
					}
				}
			}
		}
	}
	return 0, 0, 0, false
}

// RemainingPhrasesInTheme возвращает количество фраз в теме после указанной phraseID.
// Если фраза не найдена в теме, возвращает -1.
func (d *Data) RemainingPhrasesInTheme(goalID, themeID, phraseID int) int {
	goal, err := d.GetGoalByID(goalID)
	if err != nil {
		return -1
	}
	theme, err := goal.GetThemeByID(themeID)
	if err != nil {
		return -1
	}
	for i, phrase := range theme.Phrases {
		if phrase.ID == phraseID {
			return len(theme.Phrases) - i - 1
		}
	}
	return -1
}

// NextTheme возвращает ID следующей темы в той же цели, где находится themeID.
// Если themeID не найден или это последняя тема в цели — возвращает -1.
func (d *Data) NextTheme(themeID int) int {
	for _, goal := range d.Goals {
		for i, theme := range goal.Themes {
			if theme.ID == themeID {
				if i+1 < len(goal.Themes) {
					return goal.Themes[i+1].ID
				}
				return -1
			}
		}
	}
	return -2 // тема не найдена
}

// SearchPhrases ищет фразы по подстроке (в русской фразе или переводе)
func (d *Data) SearchPhrases(query string) []Phrase {
	var results []Phrase
	for _, goal := range d.Goals {
		for _, theme := range goal.Themes {
			for _, phrase := range theme.Phrases {
				if containsSubstring(phrase.Phrase, query) || containsSubstring(phrase.Translation, query) {
					results = append(results, phrase)
				}
			}
		}
	}
	return results
}

// containsSubstring проверяет, содержится ли substr в s (регистрозависимо, для простоты)
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
