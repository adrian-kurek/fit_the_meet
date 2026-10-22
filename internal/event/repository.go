package event

import (
	"context"
	"database/sql"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

type Repository struct {
	db     *sql.DB
	logger commoninterfaces.Logger
}

func NewRepository(db *sql.DB, logger commoninterfaces.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

const (
	queryMsg   = "query"
	errorMsg   = "error"
	argsMsg    = "args"
	userIDMsg  = "userID"
	groupIDMsg = "groupID"
	codeMsg    = "code"
)

func (er *Repository) Insert(ctx context.Context, userID int, code string, event Upsert) error {
	query := "INSERT INTO events(name,description, duration_minutes, code, created_by) VALUES($1,$2,$3,$4,$5)"

	stmt, err := er.db.PrepareContext(ctx, query)
	if err != nil {
		er.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			er.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, event.Name, event.Description, event.DurationMinutes, code, userID)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:         userID,
				codeMsg:           code,
				"name":            event.Name,
				"description":     event.Description,
				"durationMinutes": event.DurationMinutes,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (er *Repository) ChangeStatus(ctx context.Context, eventID, userID int) error {
	query := "UPDATE events SET is_active = NOT is_active WHERE id=$1 AND created_by = $2"

	stmt, err := er.db.PrepareContext(ctx, query)
	if err != nil {
		er.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			er.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, eventID, userID)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteUpdateQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
				"eventID": eventID,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}
