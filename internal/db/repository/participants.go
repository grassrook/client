package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"judginginternal/internal/challonge"
	"judginginternal/internal/models/participant"
	"time"
)

type ParticipantsRepo interface {
	UpdateRegisteredCar(ctx context.Context, participantID string, registeredCarID string) error
	ReplaceParticipant(ctx context.Context, from string, to string) (participant.ReplaceParticipantResult, error)
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

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getParticipant(ctx context.Context, q rowQuerier, id string) (participant.Participant, error) {
	var p participant.Participant
	var tournamentID, challongeID sql.NullString

	err := q.QueryRowContext(ctx,
		`SELECT id, name, tournament_id, challonge_participant_id FROM participants WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &tournamentID, &challongeID)
	if err != nil {
		return participant.Participant{}, err
	}

	p.TournamentID = tournamentID.String
	if challongeID.Valid {
		p.ChallongeParticipantID = &challongeID.String
	}

	return p, nil
}

func (r *SQLiteParticipantsRepository) ReplaceParticipant(ctx context.Context, from, to string) (participant.ReplaceParticipantResult, error) {
	var result participant.ReplaceParticipantResult

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE nodes SET participant_id = ? WHERE participant_id = ?`, to, from,
	); err != nil {
		return result, fmt.Errorf("update nodes.participant_id: %w", err)
	}

	// TODO: should forbid replacing participant if the participant already scored a round
	// update the scores table if it already exists
	if _, err := tx.ExecContext(ctx, `
    UPDATE scores
    SET participant_id = ?
    WHERE participant_id = ?
        AND kind = 'elimination'
`, to, from); err != nil {
		return result, fmt.Errorf("update scores.participant_id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
    UPDATE scores
    SET lead_participant_id = ?
    WHERE lead_participant_id = ?
        AND kind = 'elimination'
`, to, from); err != nil {
		return result, fmt.Errorf("update scores.lead_participant_id: %w", err)
	}

	// state of `from` before the Challonge ID is moved
	before, err := getParticipant(ctx, tx, from)
	if err != nil {
		return result, fmt.Errorf("load from participant %s: %w", from, err)
	}

	var challongeTournamentURL sql.NullString
	if before.TournamentID != "" {
		err := tx.QueryRowContext(ctx,
			`SELECT url FROM brackets WHERE tournament_id = ? LIMIT 1`, before.TournamentID,
		).Scan(&challongeTournamentURL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return result, fmt.Errorf("load bracket url for tournament %s: %w", before.TournamentID, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE participants SET challonge_participant_id = NULL WHERE id = ?`, from,
	); err != nil {
		return result, fmt.Errorf("clear challonge id on %s: %w", from, err)
	}
	if before.ChallongeParticipantID != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE participants SET challonge_participant_id = ? WHERE id = ?`,
			before.ChallongeParticipantID, to,
		); err != nil {
			return result, fmt.Errorf("set challonge id on %s: %w", to, err)
		}
	}

	// load both participants in their final state
	if result.From, err = getParticipant(ctx, tx, from); err != nil {
		return participant.ReplaceParticipantResult{}, fmt.Errorf("reload from participant %s: %w", from, err)
	}
	if result.To, err = getParticipant(ctx, tx, to); err != nil {
		return participant.ReplaceParticipantResult{}, fmt.Errorf("load to participant %s: %w", to, err)
	}

	if err := tx.Commit(); err != nil {
		return participant.ReplaceParticipantResult{}, fmt.Errorf("commit: %w", err)
	}

	if result.To.ChallongeParticipantID != nil && challongeTournamentURL.Valid && challongeTournamentURL.String != "" {
		if err := challonge.UpdateParticipantName(
			ctx, r.ChallongeAPIKey, challongeTournamentURL.String,
			*result.To.ChallongeParticipantID, result.To.Name,
		); err != nil {
			// the DB is already committed, so still return the result
			return result, fmt.Errorf("DB updated but Challonge rename failed: %w", err)
		}
	}

	return result, nil
}
