package auth

import (
	"database/sql"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

type Repository struct {
	logger commoninterfaces.Logger
	db     *sql.DB
}

func NewRepository(db *sql.DB, logger commoninterfaces.Logger) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}
