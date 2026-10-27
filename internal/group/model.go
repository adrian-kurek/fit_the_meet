package group

import "time"

type InfoWithoutParticipants struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	Role         string      `json:"role"`
	MembersCount int         `json:"member_count"`
	CreatedAt    time.Time   `json:"created_at"`
	Permissions  Permissions `json:"permissions"`
}

type Permissions struct {
	CanManageMembers bool `json:"can_manage_members"`
	CanCreateItems   bool `json:"can_create_items"`
}

type Joined struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	MembersCount int    `json:"member_count"`
}

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
