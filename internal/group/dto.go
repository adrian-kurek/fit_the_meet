package group

type Upsert struct {
	Name string `json:"username" validate:"required,min=6"`
}

type RemoveParticipants struct {
	UserIDs []int `json:"userIds" validate:"required,min=1"`
}
