package group

import (
	"context"
	"database/sql"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

const (
	queryMsg = "query"
	errorMsg = "error"
)

type Repository struct {
	logger commoninterfaces.Logger
	db     *sql.DB
}

func NewRepository(db *sql.DB, logger commoninterfaces.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (gr *Repository) Insert(ctx context.Context, name string, userID int) error {
	query := "INSERT INTO groups(name,created_by) VALUES($1,$2)"

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, name, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			"args": map[string]any{
				"userId": userID,
				"name":   name,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) UpdateName(ctx context.Context, name string, groupID, userID int) error {
	query := "UPDATE groups SET name = $1 WHERE id=$2 AND created_by = $3"

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, name, groupID, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteUpdateQuery, "data", map[string]any{
			queryMsg: query,
			"args": map[string]any{
				"userId":  userID,
				"groupID": groupID,
				"name":    name,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}
