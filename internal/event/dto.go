package event

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
