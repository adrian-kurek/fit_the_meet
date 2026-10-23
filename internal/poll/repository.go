package poll

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/lib/pq"
)

const (
	queryMsg  = "query"
	errorMsg  = "error"
	argsMsg   = "args"
	userIDMsg = "userID"
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

func (pr *Repository) Insert(ctx context.Context, userID, eventID int) error {
	const query = "INSERT INTO polls(event_id,created_by) VALUES($1,$2)"

	stmt, err := pr.db.PrepareContext(ctx, query)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			pr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, userID, eventID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
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

func (pr *Repository) InsertOptions(ctx context.Context, pollID, options []Option) error {
	args := make([]any, len(options))
	placeholders := make([]string, len(options))
	for i, option := range options {
		args = append(args, pollID, option.StartTime, option.EndTime)
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))

	}

	query := fmt.Sprintf("INSERT INTO poll_options(poll_id, start_time, end_time) VALUES %s", strings.Join(placeholders, ","))

	stmt, err := pr.db.PrepareContext(ctx, query)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			pr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				"pollID":  pollID,
				"options": options,
			},
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

	optionIDs := make([]int, 0, len(answers))
	for _, answer := range answers {
		if answer.OptionID <= 0 {
			return commonerrors.Validation("poll option IDs must be positive")
		}
		optionIDs = append(optionIDs, answer.OptionID)
	}

	const query = `WITH requested_options (option_id) AS (
		SELECT DISTINCT unnest($3::int[])
	),
	valid_options AS (
		SELECT po.id
		FROM requested_options requested
		JOIN poll_options po ON po.id = requested.option_id
		JOIN polls p ON p.id = po.poll_id
		WHERE p.id = $1
			AND p.is_active = TRUE
	),
	answers_to_insert AS (
		SELECT valid.id AS poll_option_id, $2::int AS user_id
		FROM valid_options valid
		WHERE (SELECT COUNT(*) FROM valid_options) =
			(SELECT COUNT(*) FROM requested_options)
	)
	INSERT INTO poll_answers (poll_option_id, user_id)
	SELECT poll_option_id, user_id
	FROM answers_to_insert
	ON CONFLICT (poll_option_id, user_id)
	DO UPDATE SET is_active = TRUE`

	stmt, err := pr.db.PrepareContext(ctx, query)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			pr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	result, err := stmt.ExecContext(ctx, pollID, userID, pq.Array(optionIDs))
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				"pollID":    pollID,
				userIDMsg:   userID,
				"optionIDs": optionIDs,
			},
			errorMsg: err,
		})
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return commonerrors.NotFound("active poll or poll options not found")
	}

	return nil
}
