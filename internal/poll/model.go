package poll

import "time"

type Info struct {
	ID              int
	Name            string
	Description     string
	DurationMinutes int
	Code            string
	IsActive        bool
	OptionCount     int
	ResponseCount   int
	CreatedAt       time.Time
}
