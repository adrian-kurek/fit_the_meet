package poll

import (
	"fmt"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
)

type Input struct {
	Name            string   `json:"name"             validate:"required,min=6"`
	Description     string   `json:"description"`
	DurationMinutes int      `json:"duration_minutes" validate:"required,min=1"`
	Options         []Option `json:"options"          validate:"required,min=1"`
}

type Option struct {
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time"   validate:"required"`
}

type SubmitAnswers struct {
	Answers []Answer `json:"answers" validate:"required,min=1"`
}

type Answer struct {
	OptionID int `json:"option_id" validate:"required"`
}

func (input Input) Validate() error {
	for i, option := range input.Options {
		startTime, err := parseOptionTime(option.StartTime)
		if err != nil {
			return commonerrors.Validation(fmt.Sprintf(
				"options[%d].start_time must use the RFC3339 or YYYY-MM-DDTHH:MM format",
				i,
			))
		}

		endTime, err := parseOptionTime(option.EndTime)
		if err != nil {
			return commonerrors.Validation(fmt.Sprintf(
				"options[%d].end_time must use the RFC3339 or YYYY-MM-DDTHH:MM format",
				i,
			))
		}

		if !startTime.Before(endTime) {
			return commonerrors.Validation(fmt.Sprintf(
				"options[%d].start_time must be before end_time",
				i,
			))
		}
	}

	return nil
}

func parseOptionTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05"} {
		parsedTime, err := time.Parse(layout, value)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format")
}
