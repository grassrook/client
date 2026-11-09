package repository

import (
	"context"
	"database/sql"
	tournamentactivitylog "judginginternal/internal/models/tournament_activity_log"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type TournamentActivityLogsRepository interface {
	Create(ctx context.Context, input tournamentactivitylog.CreateTournamentActivityLogInput) error
	GetByID(ctx context.Context, id string) (*tournamentactivitylog.TournamentActivityLog, error)
	GetByTournamentID(ctx context.Context, tournamentID string) ([]tournamentactivitylog.TournamentActivityLog, error)
	GetByTournamentIDAndType(ctx context.Context, tournamentID string, logType string) ([]tournamentactivitylog.TournamentActivityLog, error)
	Delete(ctx context.Context, id string) error
}

type SQLiteTournamentActivityLogsRepository struct {
	DB *sql.DB
}

func NewSQLiteTournamentActivityLogsRepository(
	db *sql.DB,
) *SQLiteTournamentActivityLogsRepository {
	return &SQLiteTournamentActivityLogsRepository{
		DB: db,
	}
}

func (r *SQLiteTournamentActivityLogsRepository) Create(
	ctx context.Context,
	input tournamentactivitylog.CreateTournamentActivityLogInput,
) error {
	now := time.Now().Unix()

	id, err := gonanoid.New(15)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO tournament_activity_logs (
			id,
			tournament_id,
			log_type,
			content,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		id,
		input.TournamentID,
		input.LogType,
		input.Content,
		now,
		now,
	)

	return err
}

func (r *SQLiteTournamentActivityLogsRepository) GetByID(
	ctx context.Context,
	id string,
) (*tournamentactivitylog.TournamentActivityLog, error) {
	var log tournamentactivitylog.TournamentActivityLog

	err := r.DB.QueryRowContext(ctx, `
		SELECT
			id,
			tournament_id,
			log_type,
			content,
			created_at,
			updated_at
		FROM tournament_activity_logs
		WHERE id = ?
	`, id).Scan(
		&log.ID,
		&log.TournamentID,
		&log.LogType,
		&log.Content,
		&log.CreatedAt,
		&log.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &log, nil
}

func (r *SQLiteTournamentActivityLogsRepository) GetByTournamentID(
	ctx context.Context,
	tournamentID string,
) ([]tournamentactivitylog.TournamentActivityLog, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			id,
			tournament_id,
			log_type,
			content,
			created_at,
			updated_at
		FROM tournament_activity_logs
		WHERE tournament_id = ?
		ORDER BY created_at DESC, id DESC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]tournamentactivitylog.TournamentActivityLog, 0)

	for rows.Next() {
		var log tournamentactivitylog.TournamentActivityLog

		if err := rows.Scan(
			&log.ID,
			&log.TournamentID,
			&log.LogType,
			&log.Content,
			&log.CreatedAt,
			&log.UpdatedAt,
		); err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *SQLiteTournamentActivityLogsRepository) GetByTournamentIDAndType(
	ctx context.Context,
	tournamentID string,
	logType string,
) ([]tournamentactivitylog.TournamentActivityLog, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			id,
			tournament_id,
			log_type,
			content,
			created_at,
			updated_at
		FROM tournament_activity_logs
		WHERE tournament_id = ?
		  AND log_type = ?
		ORDER BY created_at DESC, id DESC
	`, tournamentID, logType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]tournamentactivitylog.TournamentActivityLog, 0)

	for rows.Next() {
		var log tournamentactivitylog.TournamentActivityLog

		if err := rows.Scan(
			&log.ID,
			&log.TournamentID,
			&log.LogType,
			&log.Content,
			&log.CreatedAt,
			&log.UpdatedAt,
		); err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *SQLiteTournamentActivityLogsRepository) Delete(
	ctx context.Context,
	id string,
) error {
	_, err := r.DB.ExecContext(ctx, `
		DELETE FROM tournament_activity_logs
		WHERE id = ?
	`, id)

	return err
}
