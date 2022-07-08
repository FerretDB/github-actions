// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	if err := json.Unmarshal([]byte(settings), &iterationSettings); err != nil {
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

	if err := json.Unmarshal([]byte(settings), &singleSelectSettings); err != nil {
		return "", err
	}

	for _, option := range singleSelectSettings.Options {
		if option.ID == id {
			return option.Name, nil
		}
	}

	return "", fmt.Errorf("option id %s not found", id)
}
