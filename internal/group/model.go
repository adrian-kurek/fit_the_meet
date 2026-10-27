package group

import "time"

type Info struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Role         string        `json:"role"`
	MembersCount int           `json:"member_count"`
	CreatedAt    time.Time     `json:"created_at"`
	Participants []Participant `json:"participants"`
}

type Participant struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role"`
}
