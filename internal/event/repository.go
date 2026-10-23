package event

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	queryMsg  = "query"
	errorMsg  = "error"
	argsMsg   = "args"
	userIDMsg = "userID"
	codeMsg   = "code"

	timeBlockArgumentCount  = 5
	userIDArgumentOffset    = 1
	eventIDArgumentOffset   = 2
	dateArgumentOffset      = 3
	startTimeArgumentOffset = 4
	endTimeArgumentOffset   = 5
)

func (er *Repository) Insert(ctx context.Context, userID int, code string, event Input) error {
	const query = "INSERT INTO events(name,description, duration_minutes, code, created_by) VALUES($1,$2,$3,$4,$5)"

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

func (er *Repository) ToggleActive(ctx context.Context, eventID, userID int) error {
	const query = "UPDATE events SET is_active = NOT is_active WHERE id=$1 AND created_by = $2"

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

func (er *Repository) InsertTimeBlocks(
	ctx context.Context,
	userID, eventID int,
	timeBlocks SubmitAvailabilityRequest,
) error {
	query := `INSERT INTO event_participants(
		user_id,
		event_id,
		date,
		start_time,
		end_time
	) VALUES %s
	ON CONFLICT (
		event_id,
		user_id,
		date
	) DO UPDATE SET start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time`

	args := []any{}
	placeholders := []string{}
	for i, timeBlock := range timeBlocks.TimeBlocks {
		argumentStart := i * timeBlockArgumentCount
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d)",
			argumentStart+userIDArgumentOffset,
			argumentStart+eventIDArgumentOffset,
			argumentStart+dateArgumentOffset,
			argumentStart+startTimeArgumentOffset,
			argumentStart+endTimeArgumentOffset,
		))
		args = append(args, userID, eventID, timeBlock.Date, timeBlock.StartTime, timeBlock.EndTime)
	}
	query = fmt.Sprintf(query, strings.Join(placeholders, ","))

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

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg:  timeBlocks,
			errorMsg: err,
		})
		return err
	}

	return nil
}
