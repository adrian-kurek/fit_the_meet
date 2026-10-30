package event

import "time"

type Info struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	DurationMinutes  int       `json:"duration_minutes"`
	Code             string    `json:"code"`
	IsActive         bool      `json:"is_active"`
	CreatedBy        int       `json:"created_by"`
	GroupID          *int      `json:"group_id"`
	ResponseCount    int       `json:"response_count"`
	ParticipantCount int       `json:"participant_count"`
	CreatedAt        time.Time `json:"created_at"`
}

type Detail struct {
	Info
	TimeBlocks []TimeBlockInfo `json:"time_blocks"`
}

type TimeBlockInfo struct {
	UserID        int    `json:"user_id"`
	Username      string `json:"username"`
	Date          string `json:"date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	SubmittedByMe bool   `json:"submitted_by_me"`
}

type ListOptions struct {
	Limit    int
	Offset   int
	IsActive *bool
	GroupID  *int
}
