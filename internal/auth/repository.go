package auth

import (
	"context"
	"database/sql"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

const refreshTokenExpiration = 7 * 24 * time.Hour

type Repository struct {
	logger commoninterfaces.Logger
	db     *sql.DB
}

func NewRepository(db *sql.DB, logger commoninterfaces.Logger) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}

func (ar *Repository) InsertRefreshToken(
	ctx context.Context,
	ipAddress string,
	deviceInfo string,
	refreshToken string,
	userID int,
) error {
	query := `INSERT INTO refresh_tokens(user_id,token_hash,device_info,ip_address, expires_at, last_used_at) VALUES($1,$2,$3,$4,$5,$6)`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, closeErr)
		}
	}()

	expiration := time.Now().Add(refreshTokenExpiration)
	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, userID, refreshToken, deviceInfo, ipAddress, expiration, lastUsedAt)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteInsertQuery, map[string]any{
			"query": query,
			"args": map[string]any{
				"userId":     userID,
				"deviceInfo": deviceInfo,
				"ipAddress":  ipAddress,
			},
			"error": err,
		})
		return err
	}

	return nil
}
