package group

import (
	"context"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
)

func (gr *Repository) Insert(ctx context.Context, name string, userID int) (int, error) {
	const query = "INSERT INTO groups(name,created_by) VALUES($1,$2) RETURNING id"

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

	var id int

	err = stmt.QueryRowContext(ctx, name, userID).Scan(&id)
	if err != nil {
		gr.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]any{
				userIDMsg: userID,
				"name":    name,
			},
			errorMsg: err,
		})
		return 0, err
	}

	return id, nil
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
