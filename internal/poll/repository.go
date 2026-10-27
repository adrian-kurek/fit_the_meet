package poll

import (
	"context"
	"database/sql"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

const (
	queryMsg  = "query"
	errorMsg  = "error"
	argsMsg   = "args"
	userIDMsg = "userID"
	pollIDMsg = "pollID"

	optionArgumentCount = 3
	percentageFactor    = 100
)

type Repository struct {
	db     *sql.DB
	logger commoninterfaces.Logger
}

func NewRepository(db *sql.DB, logger commoninterfaces.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (pr *Repository) ToggleActive(ctx context.Context, pollID, userID int) error {
	const query = `UPDATE polls
	SET is_active = NOT is_active
	WHERE id = $1 AND created_by = $2`

	result, err := pr.db.ExecContext(ctx, query, pollID, userID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteUpdateQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				pollIDMsg: pollID,
				userIDMsg: userID,
			},
			errorMsg: err,
		})
		return err
	}

	return requireAffectedRow(result, "owned poll not found")
}

func (pr *Repository) Delete(ctx context.Context, userID, pollID int) error {
	const query = `DELETE FROM events e
	USING polls p
	WHERE p.id = $1
		AND p.created_by = $2
		AND e.id = p.event_id`

	result, err := pr.db.ExecContext(ctx, query, pollID, userID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return err
	}

	return requireAffectedRow(result, "owned poll not found")
}

func (pr *Repository) DeleteAnswers(ctx context.Context, userID, pollID int) error {
	const query = `WITH accessible_poll AS (
		SELECT p.id
		FROM polls p
		JOIN events e ON e.id = p.event_id
		WHERE p.id = $1
			AND (
				e.group_id IS NULL
				OR p.created_by = $2
				OR EXISTS (
					SELECT 1
					FROM group_participants gp
					WHERE gp.group_id = e.group_id AND gp.user_id = $2
				)
			)
	),
	deleted AS (
		DELETE FROM poll_answers pa
		USING poll_options po, accessible_poll p
		WHERE pa.poll_option_id = po.id
			AND po.poll_id = p.id
			AND pa.user_id = $2
		RETURNING pa.id
	)
	SELECT EXISTS(SELECT 1 FROM accessible_poll)`

	var found bool
	err := pr.db.QueryRowContext(ctx, query, pollID, userID).Scan(&found)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return err
	}
	if !found {
		return commonerrors.NotFound("poll not found")
	}

	return nil
}

func requireAffectedRow(result sql.Result, message string) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return commonerrors.NotFound(message)
	}

	return nil
}
