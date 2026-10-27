package poll

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	"github.com/lib/pq"
)

func (pr *Repository) Create(ctx context.Context, userID int, code string, poll Input) error {
	return pr.create(ctx, userID, nil, code, poll)
}

func (pr *Repository) CreateForGroup(
	ctx context.Context,
	userID, groupID int,
	code string,
	poll Input,
) error {
	return pr.create(ctx, userID, &groupID, code, poll)
}

func (pr *Repository) create(
	ctx context.Context,
	userID int,
	groupID *int,
	code string,
	poll Input,
) error {
	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			pr.logger.Error("failed to roll back poll creation", "err", rollbackErr)
		}
	}()

	eventID, err := pr.insertEvent(ctx, tx, userID, groupID, code, poll)
	if err != nil {
		return err
	}

	pollID, err := pr.insertPoll(ctx, tx, userID, eventID)
	if err != nil {
		return err
	}

	if err = pr.insertOptions(ctx, tx, pollID, poll.Options); err != nil {
		return err
	}

	return tx.Commit()
}

func (pr *Repository) insertEvent(
	ctx context.Context,
	tx *sql.Tx,
	userID int,
	groupID *int,
	code string,
	poll Input,
) (int, error) {
	const openPollQuery = `INSERT INTO events(name, description, duration_minutes, code, created_by)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id`
	const groupPollQuery = `INSERT INTO events(
		name, description, duration_minutes, code, created_by, group_id
	)
	SELECT $1, $2, $3, $4, $5, g.id
	FROM groups g
	WHERE g.id = $6 AND g.created_by = $5
	RETURNING id`

	query := openPollQuery
	args := []any{poll.Name, poll.Description, poll.DurationMinutes, code, userID}
	if groupID != nil {
		query = groupPollQuery
		args = append(args, *groupID)
	}

	var eventID int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, commonerrors.NotFound("owned group not found")
	}
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return 0, err
	}

	return eventID, nil
}

func (pr *Repository) insertPoll(ctx context.Context, tx *sql.Tx, userID, eventID int) (int, error) {
	const query = `INSERT INTO polls(event_id, created_by)
		VALUES($1, $2)
		RETURNING id`

	var pollID int
	err := tx.QueryRowContext(ctx, query, eventID, userID).Scan(&pollID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return 0, err
	}

	return pollID, nil
}

func (pr *Repository) insertOptions(ctx context.Context, tx *sql.Tx, pollID int, options []Option) error {
	if len(options) == 0 {
		return commonerrors.Validation("at least one poll option is required")
	}

	args := make([]any, 0, len(options)*optionArgumentCount)
	placeholders := make([]string, 0, len(options))
	for i, option := range options {
		argumentStart := i * optionArgumentCount
		args = append(args, pollID, option.StartTime, option.EndTime)
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d)",
			argumentStart+1,
			argumentStart+optionArgumentCount-1,
			argumentStart+optionArgumentCount,
		))
	}

	query := fmt.Sprintf(
		"INSERT INTO poll_options(poll_id, start_time, end_time) VALUES %s",
		strings.Join(placeholders, ","),
	)
	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (pr *Repository) InsertAnswers(ctx context.Context, pollID, userID int, answers []Answer) error {
	if len(answers) == 0 {
		return commonerrors.Validation("at least one poll answer is required")
	}

	uniqueOptionIDs := make(map[int]struct{}, len(answers))
	optionIDs := make([]int, 0, len(answers))
	for _, answer := range answers {
		if answer.OptionID <= 0 {
			return commonerrors.Validation("poll option IDs must be positive")
		}
		if _, exists := uniqueOptionIDs[answer.OptionID]; exists {
			continue
		}
		uniqueOptionIDs[answer.OptionID] = struct{}{}
		optionIDs = append(optionIDs, answer.OptionID)
	}

	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			pr.logger.Error("failed to roll back poll answer submission", "err", rollbackErr)
		}
	}()

	const validateQuery = `SELECT
		p.is_active,
		COUNT(DISTINCT po.id)
	FROM polls p
	JOIN events e ON e.id = p.event_id
	LEFT JOIN poll_options po ON po.poll_id = p.id AND po.id = ANY($3)
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
	GROUP BY p.is_active`

	var isActive bool
	var validOptionCount int
	err = tx.QueryRowContext(ctx, validateQuery, pollID, userID, pq.Array(optionIDs)).Scan(
		&isActive,
		&validOptionCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return commonerrors.NotFound("poll not found")
	}
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: validateQuery,
			errorMsg: err,
		})
		return err
	}
	if !isActive {
		return commonerrors.Validation("poll is inactive")
	}
	if validOptionCount != len(optionIDs) {
		return commonerrors.Validation("one or more poll options do not belong to this poll")
	}

	const deleteQuery = `DELETE FROM poll_answers pa
	USING poll_options po
	WHERE pa.poll_option_id = po.id
		AND po.poll_id = $1
		AND pa.user_id = $2`
	if _, err = tx.ExecContext(ctx, deleteQuery, pollID, userID); err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]any{
			queryMsg: deleteQuery,
			errorMsg: err,
		})
		return err
	}

	const insertQuery = `INSERT INTO poll_answers (poll_option_id, user_id)
		SELECT po.id, $2
		FROM poll_options po
		WHERE po.poll_id = $1 AND po.id = ANY($3)`
	if _, err = tx.ExecContext(ctx, insertQuery, pollID, userID, pq.Array(optionIDs)); err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: insertQuery,
			errorMsg: err,
		})
		return err
	}

	return tx.Commit()
}
