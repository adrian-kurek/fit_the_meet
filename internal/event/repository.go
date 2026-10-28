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

func (er *Repository) Insert(ctx context.Context, userID int, code string, event Input) (int, error) {
	const query = `INSERT INTO events(name, description, duration_minutes, code, created_by)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id`

	var eventID int
	err := er.db.QueryRowContext(
		ctx,
		query,
		event.Name,
		event.Description,
		event.DurationMinutes,
		code,
		userID,
	).Scan(&eventID)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return 0, err
	}

	return eventID, nil
}

func (er *Repository) InsertForGroup(ctx context.Context, userID, groupID int, code string, event Input) error {
	const query = `INSERT INTO events(
		name,
		description,
		duration_minutes,
		code,
		created_by,
		group_id
	) SELECT $1,$2,$3,$4,$5, g.id FROM groups g  WHERE g.id = $6 AND g.created_by = $5`

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

	res, err := stmt.ExecContext(ctx, event.Name, event.Description, event.DurationMinutes, code, userID, groupID)
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

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return commonerrors.NotFound("owned group not found")
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
	if len(timeBlocks.TimeBlocks) == 0 {
		return commonerrors.Validation("at least one time block is required")
	}

	query := `WITH submitted_blocks (user_id, event_id, date, start_time, end_time) AS (
		VALUES %s
	)
	INSERT INTO event_participants (
		user_id,
		event_id,
		date,
		start_time,
		end_time
	)
	SELECT
		sb.user_id,
		e.id,
		sb.date,
		sb.start_time,
		sb.end_time
	FROM submitted_blocks sb
	JOIN events e ON e.id = sb.event_id
	WHERE e.is_active = TRUE
	ON CONFLICT (event_id, user_id, date)
	DO UPDATE SET
		start_time = EXCLUDED.start_time,
		end_time = EXCLUDED.end_time`

	args := []any{}
	placeholders := []string{}
	for i, timeBlock := range timeBlocks.TimeBlocks {
		argumentStart := i * timeBlockArgumentCount
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d::date, $%d::time, $%d::time)",
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

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg:  timeBlocks,
			errorMsg: err,
		})
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return commonerrors.NotFound("active event not found")
	}

	return nil
}

func (er *Repository) GetMany(ctx context.Context, userID int, options ListOptions) ([]Info, error) {
	const query = `SELECT
		e.id,
		e.name,
		COALESCE(e.description, ''),
		e.duration_minutes,
		e.code,
		COALESCE(e.is_active, TRUE),
		e.group_id,
		(
			SELECT COUNT(DISTINCT availability.user_id)
			FROM event_participants availability
			WHERE availability.event_id = e.id
		) AS response_count,
		CASE
			WHEN e.group_id IS NULL THEN
				(
					SELECT COUNT(DISTINCT availability.user_id)
					FROM event_participants availability
					WHERE availability.event_id = e.id
				) + CASE
					WHEN EXISTS (
						SELECT 1
						FROM event_participants organizer_availability
						WHERE organizer_availability.event_id = e.id
							AND organizer_availability.user_id = e.created_by
					) THEN 0
					ELSE 1
				END
			ELSE 1 + (
				SELECT COUNT(DISTINCT member.user_id)
				FROM group_participants member
				WHERE member.group_id = e.group_id
					AND member.user_id <> e.created_by
			)
		END AS participant_count,
		e.created_at
	FROM events e
	WHERE (
		e.created_by = $1
		OR (
			e.group_id IS NOT NULL
			AND EXISTS (
				SELECT 1
				FROM group_participants membership
				WHERE membership.group_id = e.group_id
					AND membership.user_id = $1
			)
		)
		OR (
			e.group_id IS NULL
			AND EXISTS (
				SELECT 1
				FROM event_participants availability
				WHERE availability.event_id = e.id
					AND availability.user_id = $1
			)
		)
	)
	AND NOT EXISTS (
		SELECT 1
		FROM polls p
		WHERE p.event_id = e.id
	)
	AND ($2::boolean IS NULL OR e.is_active = $2)
	AND ($3::integer IS NULL OR e.group_id = $3)
	ORDER BY e.created_at DESC, e.id DESC
	LIMIT $4 OFFSET $5`

	stmt, err := er.db.PrepareContext(ctx, query)
	if err != nil {
		er.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			er.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var isActive any
	if options.IsActive != nil {
		isActive = *options.IsActive
	}
	var groupID any
	if options.GroupID != nil {
		groupID = *options.GroupID
	}

	rows, err := stmt.QueryContext(ctx, userID, isActive, groupID, options.Limit, options.Offset)
	if err != nil {
		er.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:  userID,
				"isActive": options.IsActive,
				"groupID":  options.GroupID,
				"limit":    options.Limit,
				"offset":   options.Offset,
			},
			errorMsg: err,
		})
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			er.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	events := make([]Info, 0)

	for rows.Next() {
		var event Info
		err = rows.Scan(
			&event.ID,
			&event.Name,
			&event.Description,
			&event.DurationMinutes,
			&event.Code,
			&event.IsActive,
			&event.GroupID,
			&event.ResponseCount,
			&event.ParticipantCount,
			&event.CreatedAt,
		)
		if err != nil {
			er.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err.Error(),
			})
			return nil, err
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		er.logger.Error(commonerrors.FailedToScanRows, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return nil, err
	}

	return events, nil
}
