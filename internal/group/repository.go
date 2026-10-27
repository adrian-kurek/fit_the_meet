package group

import (
	"context"
	"database/sql"
	"errors"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/lib/pq"
)

const (
	queryMsg   = "query"
	errorMsg   = "error"
	argsMsg    = "args"
	userIDMsg  = "userID"
	groupIDMsg = "groupID"
	codeMsg    = "code"
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
	const query = "INSERT INTO groups(name,created_by) VALUES($1,$2)"

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
			argsMsg: map[string]any{
				userIDMsg: userID,
				"name":    name,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) UpdateName(ctx context.Context, name string, groupID, userID int) error {
	const query = "UPDATE groups SET name = $1 WHERE id=$2 AND created_by = $3"

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
			argsMsg: map[string]any{
				userIDMsg:  userID,
				groupIDMsg: groupID,
				"name":     name,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) DeleteOwnedGroup(ctx context.Context, groupID, userID int) error {
	const query = "DELETE FROM groups WHERE id=$1 AND created_by = $2"

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

	_, err = stmt.ExecContext(ctx, groupID, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:  userID,
				groupIDMsg: groupID,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error {
	const query = `DELETE FROM group_participants gp USING groups g
		WHERE gp.group_id = g.id
		AND gp.user_id = ANY($1)
		AND g.id = $2
		AND g.created_by = $3`
	args := []any{pq.Array(participants), groupID, userID}

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

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:        userID,
				groupIDMsg:       groupID,
				"participantIDs": participants,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) VerifyOwnership(ctx context.Context, groupID, userID int) error {
	const query = `SELECT id FROM groups WHERE id = $1 AND created_by = $2`

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

	var groupIDFromDB int

	err = stmt.QueryRowContext(ctx, groupID, userID).Scan(&groupIDFromDB)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:  userID,
				groupIDMsg: groupID,
			},
			errorMsg: err.Error(),
		})
		return err
	}

	return nil
}

func (gr *Repository) SaveInviteCode(ctx context.Context, groupID int, code string, expiresAt time.Time) error {
	const query = "INSERT INTO group_activation_codes(group_id,code,expires_at) VALUES($1,$2,$3)"

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

	_, err = stmt.ExecContext(ctx, groupID, code, expiresAt)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				groupIDMsg:  groupID,
				codeMsg:     code,
				"expiresAt": expiresAt,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (gr *Repository) FindGroupIDByInviteCode(ctx context.Context, code string) (int, error) {
	const query = `SELECT group_id FROM group_activation_codes WHERE code = $1 AND expires_at > NOW()`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return 0, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var groupID int

	err = stmt.QueryRowContext(ctx, code).Scan(&groupID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				codeMsg: code,
			},
			errorMsg: err.Error(),
		})
		return 0, err
	}

	return groupID, nil
}

func (gr *Repository) AddMember(ctx context.Context, groupID, userID int) error {
	const query = `INSERT INTO group_participants (group_id, user_id)
		SELECT g.id, $2
		FROM groups g
		WHERE g.id = $1 AND g.created_by <> $2
		ON CONFLICT (group_id, user_id) DO NOTHING`

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

	_, err = stmt.ExecContext(ctx, groupID, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg:  userID,
				groupIDMsg: groupID,
			},
			errorMsg: err.Error(),
		})
		return err
	}

	return nil
}

func (gr *Repository) GetGroupNamesForUser(ctx context.Context, userID int) ([]string, error) {
	const query = `
	SELECT 
		name 
	FROM groups g 
	INNER JOIN group_participants gp ON gp.group_id = g.id 
	WHERE g.created_by  = $1 OR gp.user_id = $1`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var names []string

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
			},
			errorMsg: err.Error(),
		})
		return nil, err
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			gr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err.Error(),
			})
			return nil, err
		}
		names = append(names, name)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	return names, nil
}

func (gr *Repository) GetEventsForTheGroup(ctx context.Context, groupID, userID int) ([]string, error) {
	const query = `
	SELECT 
		name 
	FROM groups g 
	INNER JOIN group_participants gp ON gp.group_id = g.id 
	WHERE g.created_by  = $1 OR gp.user_id = $1`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var names []string

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
			},
			errorMsg: err.Error(),
		})
		return nil, err
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			gr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err.Error(),
			})
			return nil, err
		}
		names = append(names, name)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	return names, nil
}

func (gr *Repository) GetGroups(ctx context.Context, userID int) ([]string, error) {
	const query = `
	SELECT
		g.id,
		g.name,
		IF(g.created_by = $1,"owner","member") as role,
		jmc.members_count,
		g.created_at
	FROM groups g
	INNER JOIN group_participants gp ON gp.group_id = g.id
	INNER JOIN (
			SELECT g.id,COUNT(*) as members_count FROM groups g
			INNER JOIN group_participants gp ON gp.group_id = g.id
			GROUP BY  g.id
		) as join_member_count jmc ON  jmc.id = g.id
	WHERE g.created_by  = $1 OR gp.user_id = $1`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var names []string

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
			},
			errorMsg: err.Error(),
		})
		return nil, err
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			gr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err.Error(),
			})
			return nil, err
		}
		names = append(names, name)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	return names, nil
}

func (gr *Repository) getInfoAboutParticipants(
	ctx context.Context,
	groupID int,
	includeEmail bool,
) ([]Participant, error) {
	const query = `WITH group_members AS (
		SELECT g.created_by AS user_id, 'owner'::text AS role
		FROM groups g
		WHERE g.id = $1

		UNION ALL

		SELECT DISTINCT gp.user_id, 'member'::text AS role
		FROM group_participants gp
		JOIN groups g ON g.id = gp.group_id
		WHERE gp.group_id = $1 AND gp.user_id <> g.created_by
	)
	SELECT
		u.id,
		u.username,
		CASE WHEN $2 THEN u.email ELSE NULL END,
		gm.role
	FROM group_members gm
	JOIN users u ON u.id = gm.user_id
	ORDER BY CASE WHEN gm.role = 'owner' THEN 0 ELSE 1 END, u.username`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	participants := make([]Participant, 0)

	rows, err := stmt.QueryContext(ctx, groupID, includeEmail)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				groupIDMsg: groupID,
			},
			errorMsg: err.Error(),
		})
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseRows, "err", closeErr)
		}
	}()

	for rows.Next() {
		var participant Participant
		var email sql.NullString
		err = rows.Scan(&participant.ID, &participant.Username, &email, &participant.Role)
		if err != nil {
			gr.logger.Error(commonerrors.FailedToScanRow, "data", map[string]any{
				queryMsg: query,
				errorMsg: err.Error(),
			})
			return nil, err
		}
		if email.Valid {
			participant.Email = email.String
		}
		participants = append(participants, participant)
	}
	if err = rows.Err(); err != nil {
		gr.logger.Error(commonerrors.FailedToScanRows, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return nil, err
	}

	return participants, nil
}

func (gr *Repository) GetByID(ctx context.Context, userID, groupID int) (Info, error) {
	const query = `SELECT
		g.id,
		g.name,
		CASE WHEN g.created_by = $1 THEN 'owner' ELSE 'member' END,
		1 + COUNT(DISTINCT gp.user_id) FILTER (WHERE gp.user_id <> g.created_by),
		g.created_at
	FROM groups g
	LEFT JOIN group_participants gp ON gp.group_id = g.id
	WHERE g.id = $2
		AND (
			g.created_by = $1
			OR EXISTS (
				SELECT 1
				FROM group_participants membership
				WHERE membership.group_id = g.id AND membership.user_id = $1
			)
		)
	GROUP BY g.id, g.name, g.created_by, g.created_at`

	stmt, err := gr.db.PrepareContext(ctx, query)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return Info{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			gr.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	var group Info

	err = stmt.QueryRowContext(ctx, userID, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.Role,
		&group.MembersCount,
		&group.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Info{}, commonerrors.NotFound("group not found")
	}
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
			},
			errorMsg: err.Error(),
		})
		return Info{}, err
	}

	participants, err := gr.getInfoAboutParticipants(ctx, groupID, group.Role == "owner")
	if err != nil {
		return Info{}, err
	}

	group.Participants = participants

	return group, nil
}
