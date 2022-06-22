package graphql

import (
	"encoding/json"
	"fmt"
)

// IterationSettings stores settings for ITERATION field.
type IterationSettings struct {
	Configuration struct {
		Duration            int                    `json:"duration"`
		StartDay            int                    `json:"start_day"`
		Iterations          []IterationSettingsOne `json:"iterations"`
		CompletedIterations []IterationSettingsOne `json:"completed_iterations"`
	} `json:"configuration"`
}

// IterationSettingsOne represents a single iteration in the iteration settings.
type IterationSettingsOne struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Duration  int    `json:"duration"`
	StartDate string `json:"start_date"`
	TitleHTML string `json:"title_html"`
}

// SingleSelectSettings stores settings for SINGLE_SELECT field.
type SingleSelectSettings struct {
	Options []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		NameHTML string `json:"name_html"`
	} `json:"options"`
}

// GetIterationTitleByID returns the title of the iteration with the given id.
// It receives iteration settings stored as a string and decodes them.
func GetIterationTitleByID(id string, settings string) (string, error) {
	var iterationSettings IterationSettings

	err := json.Unmarshal([]byte(settings), &iterationSettings)
	if err != nil {
		return "", err
	}

	all := append(iterationSettings.Configuration.Iterations, iterationSettings.Configuration.CompletedIterations...)
	for _, iteration := range all {
		if iteration.ID == id {
			return iteration.Title, nil
		}
	}

	return "", fmt.Errorf("iteration id %s not found", id)
}

// GetSingleSelectTitleByID returns the title of the option with the given id.
// It receives single select settings stored as a string and decodes them.
func GetSingleSelectTitleByID(id string, settings string) (string, error) {
	var singleSelectSettings SingleSelectSettings

	err := json.Unmarshal([]byte(settings), &singleSelectSettings)
	if err != nil {
		return "", err
	}

	for _, option := range singleSelectSettings.Options {
		if option.ID == id {
			return option.Name, nil
		}
	}

	return "", fmt.Errorf("option id %s not found", id)
}
