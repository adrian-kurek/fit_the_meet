package poll

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
