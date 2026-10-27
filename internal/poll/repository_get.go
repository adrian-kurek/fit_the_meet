package poll

import (
	"context"
	"database/sql"
	"errors"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
)

func (pr *Repository) Get(ctx context.Context, userID int) ([]Info, error) {
	const query = `SELECT
		p.id,
		p.event_id,
		e.name,
		e.description,
		e.duration_minutes,
		e.code,
		p.is_active,
		e.group_id,
		(SELECT COUNT(*) FROM poll_options po WHERE po.poll_id = p.id),
		(
			SELECT COUNT(DISTINCT pa.user_id)
			FROM poll_answers pa
			JOIN poll_options po ON po.id = pa.poll_option_id
			WHERE po.poll_id = p.id AND pa.is_active = TRUE
		),
		p.created_at
	FROM polls p
	JOIN events e ON e.id = p.event_id
	WHERE p.created_by = $1
		OR EXISTS (
			SELECT 1
			FROM group_participants gp
			WHERE gp.group_id = e.group_id AND gp.user_id = $1
		)
		OR EXISTS (
			SELECT 1
			FROM poll_answers pa
			JOIN poll_options po ON po.id = pa.poll_option_id
			WHERE po.poll_id = p.id AND pa.user_id = $1
		)
	ORDER BY p.created_at DESC`

	rows, err := pr.db.QueryContext(ctx, query, userID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	polls := make([]Info, 0)
	for rows.Next() {
		var poll Info
		if err = rows.Scan(
			&poll.ID,
			&poll.EventID,
			&poll.Name,
			&poll.Description,
			&poll.DurationMinutes,
			&poll.Code,
			&poll.IsActive,
			&poll.GroupID,
			&poll.OptionCount,
			&poll.ResponseCount,
			&poll.CreatedAt,
		); err != nil {
			pr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err,
			})
			return nil, err
		}
		polls = append(polls, poll)
	}
	if err = rows.Err(); err != nil {
		pr.logger.Error(commonerrors.FailedToScanRows, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return nil, err
	}

	return polls, nil
}

func (pr *Repository) GetByID(ctx context.Context, userID, pollID int) (Detail, error) {
	poll, err := pr.getInfoByID(ctx, userID, pollID)
	if err != nil {
		return Detail{}, err
	}

	poll.Options, err = pr.getOptions(ctx, userID, pollID, poll.OptionCount)
	if err != nil {
		return Detail{}, err
	}

	return poll, nil
}

func (pr *Repository) getInfoByID(ctx context.Context, userID, pollID int) (Detail, error) {
	const query = `SELECT
		p.id,
		p.event_id,
		e.name,
		e.description,
		e.duration_minutes,
		e.code,
		p.is_active,
		e.group_id,
		(SELECT COUNT(*) FROM poll_options po WHERE po.poll_id = p.id),
		(
			SELECT COUNT(DISTINCT pa.user_id)
			FROM poll_answers pa
			JOIN poll_options po ON po.id = pa.poll_option_id
			WHERE po.poll_id = p.id AND pa.is_active = TRUE
		),
		p.created_at,
		CASE
			WHEN e.group_id IS NULL THEN (
				SELECT COUNT(DISTINCT pa.user_id)
				FROM poll_answers pa
				JOIN poll_options po ON po.id = pa.poll_option_id
				WHERE po.poll_id = p.id AND pa.is_active = TRUE
			)
			ELSE 1 + (
				SELECT COUNT(DISTINCT gp.user_id)
				FROM group_participants gp
				WHERE gp.group_id = e.group_id
			)
		END
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
		)`

	var poll Detail
	err := pr.db.QueryRowContext(ctx, query, pollID, userID).Scan(
		&poll.ID,
		&poll.EventID,
		&poll.Name,
		&poll.Description,
		&poll.DurationMinutes,
		&poll.Code,
		&poll.IsActive,
		&poll.GroupID,
		&poll.OptionCount,
		&poll.ResponseCount,
		&poll.CreatedAt,
		&poll.ParticipantsCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Detail{}, commonerrors.NotFound("poll not found")
	}
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return Detail{}, err
	}

	return poll, nil
}

func (pr *Repository) getOptions(
	ctx context.Context,
	userID, pollID, optionCount int,
) ([]OptionInfo, error) {
	const optionsQuery = `SELECT
		po.id,
		po.start_time,
		po.end_time,
		COUNT(DISTINCT pa.user_id),
		COALESCE(BOOL_OR(pa.user_id = $2 AND pa.is_active = TRUE), FALSE)
	FROM poll_options po
	LEFT JOIN poll_answers pa ON pa.poll_option_id = po.id AND pa.is_active = TRUE
	WHERE po.poll_id = $1
	GROUP BY po.id, po.start_time, po.end_time
	ORDER BY po.start_time`

	rows, err := pr.db.QueryContext(ctx, optionsQuery, pollID, userID)
	if err != nil {
		pr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: optionsQuery,
			errorMsg: err,
		})
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	options := make([]OptionInfo, 0, optionCount)
	for rows.Next() {
		var option OptionInfo
		if err = rows.Scan(
			&option.ID,
			&option.StartAt,
			&option.EndAt,
			&option.VoteCount,
			&option.SelectedByMe,
		); err != nil {
			pr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: optionsQuery,
				errorMsg: err,
			})
			return nil, err
		}
		options = append(options, option)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func (pr *Repository) GetResults(ctx context.Context, userID, pollID int) (Result, error) {
	poll, err := pr.GetByID(ctx, userID, pollID)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		ID:                poll.ID,
		ParticipantsCount: poll.ParticipantsCount,
		ResponseCount:     poll.ResponseCount,
		Options:           make([]OptionResult, 0, len(poll.Options)),
	}
	for _, option := range poll.Options {
		percentage := 0
		if poll.ResponseCount > 0 {
			percentage = option.VoteCount * percentageFactor / poll.ResponseCount
		}
		result.Options = append(result.Options, OptionResult{
			ID:         option.ID,
			StartAt:    option.StartAt,
			EndAt:      option.EndAt,
			VoteCount:  option.VoteCount,
			Percentage: percentage,
		})
	}

	return result, nil
}

