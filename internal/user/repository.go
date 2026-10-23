package user

import (
	"context"
	"database/sql"
	"errors"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

const (
	errorMsg = "error"
	queryMsg = "query"
	emailMsg = "email"
	argsMsg  = "args"
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

func (ur *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = "SELECT id FROM users WHERE email = $1"

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return false, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	doesExist := true
	var id int

	err = stmt.QueryRowContext(ctx, email).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			doesExist = false
		} else {
			ur.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
				queryMsg: query,
				argsMsg: map[string]string{
					emailMsg: email,
				},
				errorMsg: err,
			})
			return false, err
		}
	}

	return doesExist, nil
}

func (ur *Repository) Insert(ctx context.Context, username, email, password string) error {
	const query = "INSERT INTO users(username,email,password) VALUES($1,$2,$3)"

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, username, email, password)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]string{
				"username": username,
				emailMsg:   email,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (ur *Repository) GetByEmail(ctx context.Context, email string) (Account, error) {
	const query = "SELECT id,email,username,password FROM users WHERE email = $1"

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return Account{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var user Account

	err = stmt.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, commonerrors.NotFound("failed to find user with provided email")
		}
		ur.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]string{
				"email": email,
			},
			errorMsg: err,
		})
		return Account{}, err
	}

	return user, nil
}
