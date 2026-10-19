package user

import (
	"context"
	"database/sql"
	"errors"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/internal/auth"
)

const (
	errorMsg = "error"
	queryMsg = "query"
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

func (ur *Repository) CheckDoesExistByEmail(ctx context.Context, email string) (bool, error) {
	query := "SELECT id FROM users WHERE email = $1"

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToPrepareQuery, map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return false, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.logger.Error(commonerrors.FailedToCloseStatement, closeErr)
		}
	}()

	doesExist := true
	var id int

	err = stmt.QueryRowContext(ctx, email).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			doesExist = false
		} else {
			ur.logger.Error(commonerrors.FailedToExecuteSelectQuery, map[string]any{
				queryMsg: query,
				"args": map[string]string{
					"email": email,
				},
				errorMsg: err,
			})
			return false, err
		}
	}

	return doesExist, nil
}

func (ur *Repository) Insert(ctx context.Context, register auth.Register) error {
	query := "INSERT INTO users(username,email,password) VALUES($1,$2,$3)"

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToPrepareQuery, map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.logger.Error(commonerrors.FailedToCloseStatement, closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, register.Username, register.Email, register.Password)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToExecuteInsertQuery, map[string]any{
			queryMsg: query,
			"args": map[string]string{
				"username": register.Username,
				"email":    register.Email,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}
