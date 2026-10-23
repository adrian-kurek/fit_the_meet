package group

type Input struct {
	Name string `json:"username" validate:"required,min=6"`
}

type RemoveParticipantsRequest struct {
	UserIDs []int `json:"userIds" validate:"required,min=1"`
}
