package event

import "time"

type Info struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	DurationMinutes  int       `json:"duration_minutes"`
	Code             string    `json:"code"`
	IsActive         bool      `json:"is_active"`
	GroupID          *int      `json:"group_id"`
	ResponseCount    int       `json:"response_count"`
	ParticipantCount int       `json:"participant_count"`
	CreatedAt        time.Time `json:"created_at"`
}

type ListOptions struct {
	Limit    int
	Offset   int
	IsActive *bool
	GroupID  *int
}
