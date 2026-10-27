package group

import "time"

type Event struct {
	ID          int
	Name        string
	Description string
	// duration_minutes
}

type Info struct {
	ID           int
	Name         string
	Role         string
	MembersCount int
	CreatedAt    time.Time
}
