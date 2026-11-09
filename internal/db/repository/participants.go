package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"judginginternal/internal/challonge"
	"time"
)

type ParticipantsRepo interface {
	UpdateRegisteredCar(ctx context.Context, participantID string, registeredCarID string) error
	ReplaceParticipant(ctx context.Context, from string, to string) error
}

type SQLiteParticipantsRepository struct {
	DB              *sql.DB
	ChallongeAPIKey string
}

func (r *SQLiteParticipantsRepository) UpdateRegisteredCar(ctx context.Context, participantID string, registeredCarID string) error {
	var carID sql.NullString

	if registeredCarID != "" {
		carID = sql.NullString{
			String: registeredCarID,
			Valid:  true,
		}
	}

	_, err := r.DB.ExecContext(ctx, `
		UPDATE participants
		SET registered_car_id = ?, updated_at = ?
		WHERE id = ?
	`, carID, time.Now().Unix(), participantID)

	return err
}

func (r *SQLiteParticipantsRepository) ReplaceParticipant(ctx context.Context, from, to string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE nodes SET participant_id = ? WHERE participant_id = ?`, to, from,
	); err != nil {
		return fmt.Errorf("update nodes.participant_id: %w", err)
	}

	var challongeID, tournamentID sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT challonge_participant_id, tournament_id FROM participants WHERE id = ?`, from,
	).Scan(&challongeID, &tournamentID); err != nil {
		return fmt.Errorf("load from participant %s: %w", from, err)
	}

	var challongeTournamentURL sql.NullString
	if tournamentID.Valid {
		err := tx.QueryRowContext(ctx,
			`SELECT url FROM brackets WHERE tournament_id = ? LIMIT 1`, tournamentID.String,
		).Scan(&challongeTournamentURL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("load bracket url for tournament %s: %w", tournamentID.String, err)
		}
	}

	var toName string
	if err := tx.QueryRowContext(ctx,
		`SELECT name FROM participants WHERE id = ?`, to,
	).Scan(&toName); err != nil {
		return fmt.Errorf("load to participant %s: %w", to, err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE participants SET challonge_participant_id = NULL WHERE id = ?`, from,
	); err != nil {
		return fmt.Errorf("clear challonge id on %s: %w", from, err)
	}
	if challongeID.Valid {
		if _, err := tx.ExecContext(ctx,
			`UPDATE participants SET challonge_participant_id = ? WHERE id = ?`, challongeID.String, to,
		); err != nil {
			return fmt.Errorf("set challonge id on %s: %w", to, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if challongeID.Valid && challongeTournamentURL.Valid && challongeTournamentURL.String != "" {
		if err := challonge.UpdateParticipantName(ctx, r.ChallongeAPIKey, challongeTournamentURL.String, challongeID.String, toName); err != nil {
			return fmt.Errorf("DB updated but Challonge rename failed: %w", err)
		}
	}

	return nil
}
