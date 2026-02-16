package database

import (
	"aura/logging"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
)

func (s *SQliteDB) CreateAuthTable(ctx context.Context) (Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Creating AUTH Table", logging.LevelDebug)
	defer logAction.Complete()

	query := `
	CREATE TABLE IF NOT EXISTS AUTH (
		token_secret TEXT NOT NULL
	);
	`
	_, err := s.conn.ExecContext(ctx, query)
	if err != nil {
		logAction.SetError("Failed to create AUTH table", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return *logAction.Error
	}

	return logging.LogErrorInfo{}
}

func (s *SQliteDB) GetAuthTokenSecret(ctx context.Context) (secret string, Err logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Getting Auth Token Secret", logging.LevelDebug)
	defer logAction.Complete()

	query := `SELECT token_secret FROM AUTH LIMIT 1;`
	row := s.conn.QueryRowContext(ctx, query)

	var tokenSecret string
	err := row.Scan(&tokenSecret)
	if err != nil {
		// If no secret exists yet, generate + persist one.
		if err == sql.ErrNoRows {
			newSecret, genErr := generateTokenAuthSecret()
			if genErr != nil {
				logAction.SetError("Failed to generate Auth Token Secret", genErr.Error(), map[string]any{
					"error": genErr.Error(),
				})
				return "", *logAction.Error
			}

			insertQuery := `INSERT INTO AUTH (token_secret) VALUES (?);`
			if _, execErr := s.conn.ExecContext(ctx, insertQuery, newSecret); execErr != nil {
				logAction.SetError("Failed to persist Auth Token Secret", execErr.Error(), map[string]any{
					"error": execErr.Error(),
					"query": insertQuery,
				})
				return "", *logAction.Error
			}

			return newSecret, logging.LogErrorInfo{}
		}

		logAction.SetError("Failed to get Auth Token Secret", err.Error(), map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return "", *logAction.Error
	}

	return tokenSecret, logging.LogErrorInfo{}
}

// generateTokenAuthSecret generates a random 256-bit secret suitable for HS256.
func generateTokenAuthSecret() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe, no padding; good for env vars/files.
	return base64.RawURLEncoding.EncodeToString(b), nil
}
