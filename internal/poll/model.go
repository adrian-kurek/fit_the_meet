package poll

import "time"

type Info struct {
	ID              int       `json:"id"`
	EventID         int       `json:"event_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Code            string    `json:"code"`
	IsActive        bool      `json:"is_active"`
	GroupID         *int      `json:"group_id"`
	OptionCount     int       `json:"option_count"`
	ResponseCount   int       `json:"response_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type Result struct {
	ID                int            `json:"poll_id"`
	ParticipantsCount int            `json:"participant_count"`
	ResponseCount     int            `json:"response_count"`
	Options           []OptionResult `json:"options"`
}

type OptionResult struct {
	ID         int       `json:"id"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	VoteCount  int       `json:"vote_count"`
	Percentage int       `json:"percentage"`
}

type Detail struct {
	Info

	ParticipantsCount int          `json:"participant_count"`
	Options           []OptionInfo `json:"options"`
}

type OptionInfo struct {
	ID           int       `json:"id"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	VoteCount    int       `json:"vote_count"`
	SelectedByMe bool      `json:"selected_by_me"`
}

type CreatedOption struct {
	ID      int       `json:"id"`
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
}

type Created struct {
	ID       int             `json:"id"`
	EventID  int             `json:"event_id"`
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	IsActive bool            `json:"is_active"`
	GroupID  *int            `json:"group_id"`
	Options  []CreatedOption `json:"options"`
}
