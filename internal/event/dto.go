package event

import (
	"fmt"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
)

const timeOfDayLayout = "15:04"

type Input struct {
	Name            string `json:"name"             validate:"required,min=6"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes" validate:"required,min=1"`
}

type SubmitAvailabilityRequest struct {
	TimeBlocks []TimeBlock `json:"time_blocks" validate:"required,min=1"`
}

type TimeBlock struct {
	Date      string `json:"date"       validate:"required"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time"   validate:"required"`
}

func (request SubmitAvailabilityRequest) Validate() error {
	for i, block := range request.TimeBlocks {
		if _, err := time.Parse(time.DateOnly, block.Date); err != nil {
			return commonerrors.Validation(fmt.Sprintf(
				"time_blocks[%d].date must use the YYYY-MM-DD format",
				i,
			))
		}

		startTime, err := time.Parse(timeOfDayLayout, block.StartTime)
		if err != nil {
			return commonerrors.Validation(fmt.Sprintf(
				"time_blocks[%d].start_time must use the HH:MM format",
				i,
			))
		}

		endTime, err := time.Parse(timeOfDayLayout, block.EndTime)
		if err != nil {
			return commonerrors.Validation(fmt.Sprintf(
				"time_blocks[%d].end_time must use the HH:MM format",
				i,
			))
		}

		if !startTime.Before(endTime) {
			return commonerrors.Validation(fmt.Sprintf(
				"time_blocks[%d].start_time must be before end_time",
				i,
			))
		}
	}

	return nil
}
