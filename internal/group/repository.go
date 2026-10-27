package group

import (
	"context"
	"database/sql"

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
