package event

type Upsert struct {
	Name            string `json:"name"             validate:"required,min=6"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes" validate:"required,min=1"`
}
