// Package prompter provides user interaction adapters.
package prompter

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/keyxmare/DevBootstrap/internal/port/secondary"
)

// SurveyPrompter implements UserPrompter using the survey library.
type SurveyPrompter struct {
	noInteraction bool
	reporter      secondary.ProgressReporter
}

// NewSurveyPrompter creates a new SurveyPrompter instance.
func NewSurveyPrompter(noInteraction bool, reporter secondary.ProgressReporter) *SurveyPrompter {
	return &SurveyPrompter{
		noInteraction: noInteraction,
		reporter:      reporter,
	}
}

// Confirm asks a yes/no question.
func (p *SurveyPrompter) Confirm(question string, defaultValue bool) bool {
	if p.noInteraction {
		if p.reporter != nil {
			answer := "non"
			if defaultValue {
				answer = "oui"
			}
			p.reporter.Info(question + " → " + answer + " (auto)")
		}
		return defaultValue
	}

	result := defaultValue
	prompt := &survey.Confirm{
		Message: question,
		Default: defaultValue,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultValue
	}

	return result
}

// Select asks user to select one option.
func (p *SurveyPrompter) Select(question string, options []string, defaultIndex int) int {
	if p.noInteraction {
		if p.reporter != nil {
			p.reporter.Info(question + " → " + options[defaultIndex] + " (auto)")
		}
		return defaultIndex
	}

	var result string
	prompt := &survey.Select{
		Message: question,
		Options: options,
		Default: options[defaultIndex],
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultIndex
	}

	for i, opt := range options {
		if opt == result {
			return i
		}
	}

	return defaultIndex
}

// MultiSelect asks user to select multiple options.
func (p *SurveyPrompter) MultiSelect(question string, options []string, defaultIndices []int) []int {
	if p.noInteraction {
		if p.reporter != nil {
			p.reporter.Info(question + " → selection automatique")
		}
		return defaultIndices
	}

	var results []string
	defaultStrs := make([]string, 0, len(defaultIndices))
	for _, idx := range defaultIndices {
		if idx < len(options) {
			defaultStrs = append(defaultStrs, options[idx])
		}
	}

	prompt := &survey.MultiSelect{
		Message: question,
		Options: options,
		Default: defaultStrs,
	}

	if err := survey.AskOne(prompt, &results); err != nil {
		return defaultIndices
	}

	indices := make([]int, 0, len(results))
	for _, result := range results {
		for i, opt := range options {
			if opt == result {
				indices = append(indices, i)
				break
			}
		}
	}

	return indices
}

// Input asks user for text input.
func (p *SurveyPrompter) Input(question string, defaultValue string) string {
	if p.noInteraction {
		if p.reporter != nil {
			p.reporter.Info(question + " → " + defaultValue + " (auto)")
		}
		return defaultValue
	}

	var result string
	prompt := &survey.Input{
		Message: question,
		Default: defaultValue,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return defaultValue
	}

	return result
}

// Password asks user for password input (hidden).
func (p *SurveyPrompter) Password(question string) string {
	if p.noInteraction {
		return ""
	}

	var result string
	prompt := &survey.Password{
		Message: question,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return ""
	}

	return result
}
