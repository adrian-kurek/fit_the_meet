package auth

import (
	"context"
	"database/sql"
	"errors"
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

const (
	queryMsg = "query"
	errorMsg = "error"
)

func (ar *Repository) SaveRefreshToken(
	ctx context.Context,
	ipAddress string,
	deviceInfo string,
	refreshToken string,
	userID int,
) error {
	const query = `INSERT INTO refresh_tokens(user_id,token_hash,device_info,ip_address, expires_at, last_used_at) VALUES($1,$2,$3,$4,$5,$6)`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	expiration := time.Now().Add(refreshTokenExpiration)
	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, userID, refreshToken, deviceInfo, ipAddress, expiration, lastUsedAt)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			"args": map[string]any{
				"userId":     userID,
				"deviceInfo": deviceInfo,
				"ipAddress":  ipAddress,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (ar *Repository) GetRefreshTokenByTokenHash(ctx context.Context, tokenHash string) (TokenWithUser, error) {
	const query = `
	SELECT 
		rt.id,
		rt.user_id,
		u.email,
		u.username,
		rt.token_hash,
		rt.expires_at 
	FROM refresh_tokens rt
		INNER JOIN users u ON u.id = rt.user_id
	WHERE
		rt.token_hash = $1`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return TokenWithUser{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	row := stmt.QueryRowContext(ctx, tokenHash)

	var tokenWithUser TokenWithUser
	err = row.Scan(
		&tokenWithUser.ID,
		&tokenWithUser.UserID,
		&tokenWithUser.Email,
		&tokenWithUser.Username,
		&tokenWithUser.TokenHash,
		&tokenWithUser.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ar.logger.Info("token not found")
			return TokenWithUser{ID: 0}, nil
		}
		ar.logger.Error(commonerrors.FailedToExecuteSelectQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return TokenWithUser{}, err
	}
	return tokenWithUser, nil
}

func (ar *Repository) UpdateLastTimeUsedToken(ctx context.Context, refreshToken string) error {
	const query = `
	UPDATE refresh_tokens SET last_used_at = $1 WHERE token_hash = $2	
	`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, "data", closeErr)
		}
	}()

	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, lastUsedAt, refreshToken)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteUpdateQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}

	return nil
}

func (ar *Repository) RemoveTokenByToken(ctx context.Context, refreshToken string) error {
	const query = "DELETE FROM refresh_tokens WHERE token_hash = $1"

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, refreshToken)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}

	return nil
}

func (ar *Repository) RemoveTokensByUserID(ctx context.Context, userID int) error {
	const query = "DELETE FROM refresh_tokens WHERE user_id = $1"

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.logger.Error(commonerrors.FailedToCloseStatement, "err", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, userID)
	if err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteDeleteQuery, "data", map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}

	return nil
}

func (ar *Repository) SavePasswordResetToken(
	ctx context.Context,
	userID int,
	tokenHash string,
	expiresAt time.Time,
) error {
	const query = `INSERT INTO password_reset_tokens(user_id, token_hash, expires_at)
	VALUES($1, $2, $3)
	ON CONFLICT (user_id) DO UPDATE SET
		token_hash = EXCLUDED.token_hash,
		expires_at = EXCLUDED.expires_at,
		created_at = CURRENT_TIMESTAMP`

	if _, err := ar.db.ExecContext(ctx, query, userID, tokenHash, expiresAt); err != nil {
		ar.logger.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (ar *Repository) ResetPassword(ctx context.Context, tokenHash, password string) error {
	tx, err := ar.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			ar.logger.Error("failed to roll back password reset", "err", rollbackErr)
		}
	}()

	const tokenQuery = `SELECT user_id
	FROM password_reset_tokens
	WHERE token_hash = $1 AND expires_at > NOW()
	FOR UPDATE`
	var userID int
	if err = tx.QueryRowContext(ctx, tokenQuery, tokenHash).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return commonerrors.Validation("password reset link is invalid or has expired")
		}
		return err
	}

	const updatePasswordQuery = "UPDATE users SET password = $1 WHERE id = $2"
	if _, err = tx.ExecContext(ctx, updatePasswordQuery, password, userID); err != nil {
		return err
	}

	const removeSessionsQuery = "DELETE FROM refresh_tokens WHERE user_id = $1"
	if _, err = tx.ExecContext(ctx, removeSessionsQuery, userID); err != nil {
		return err
	}

	const removeResetTokensQuery = "DELETE FROM password_reset_tokens WHERE user_id = $1"
	if _, err = tx.ExecContext(ctx, removeResetTokensQuery, userID); err != nil {
		return err
	}

	return tx.Commit()
}
