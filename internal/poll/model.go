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

type Result struct {
	ID                int
	ParticipantsCount int
	ResponseCount     int
	options           []OptionResult
}

type OptionResult struct {
	ID         int
	StartAt    time.Time
	EndAt      time.Time
	VoteCount  int
	Percentage int
}

type CreatedOption struct {
	ID        int
	StartedAt time.Time
	EndAt     time.Time
}

type Created struct {
	ID       int
	EventID  int
	Code     string
	Name     string
	IsActive bool
	GroupID  int
	Options  []CreatedOption
}
