package group

type Upsert struct {
	Name string `json:"username" validate:"required,min=6"`
}
