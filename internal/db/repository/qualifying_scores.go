package repository

import (
	"context"
	"database/sql"
	"fmt"
	qualifyingscore "judginginternal/internal/models/qualifying_score"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type QualfyingScoresRepository interface {
	FindAllScores(ctx context.Context, tournamentID string) ([]qualifyingscore.Data, error)
	FindScoresByParticipantID(
		ctx context.Context,
		tournamentID string,
		participantID string,
	) ([]qualifyingscore.QualifyingScore, error)
	InitializeScores(ctx context.Context, tournamentID string) error
	UpdateScore(ctx context.Context, tournamentID, participantID string, round int, score float64, status string) error
}

type SQLiteQualifyingScoresRepository struct {
	DB *sql.DB
}

func (r *SQLiteQualifyingScoresRepository) FindAllScores(
	ctx context.Context,
	tournamentID string,
) ([]qualifyingscore.Data, error) {
	var qualifyingRounds int

	err := r.DB.QueryRowContext(ctx, `
        SELECT qualifying_rounds
        FROM tournaments
        WHERE id = ?
    `, tournamentID).Scan(&qualifyingRounds)
	if err != nil {
		return nil, fmt.Errorf("get qualifying rounds: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, `
        SELECT p.id, qs.round, qs.score, qs.status
        FROM participants p
        LEFT JOIN qualifying_scores qs
            ON qs.participant_id = p.id
            AND qs.tournament_id = ?
        WHERE p.tournament_id = ?
        ORDER BY p.id, qs.round
    `, tournamentID, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("query qualifying scores: %w", err)
	}
	defer rows.Close()

	participants := make(map[string]*qualifyingscore.Data)
	var result []qualifyingscore.Data

	for rows.Next() {
		var participantID string
		var round sql.NullInt64
		var score sql.NullFloat64
		var status string

		if err := rows.Scan(&participantID, &round, &score, &status); err != nil {
			return nil, fmt.Errorf("scan qualifying score: %w", err)
		}

		data, exists := participants[participantID]
		if !exists {
			data = &qualifyingscore.Data{
				ParticipantID: participantID,
				Scores:        make([]qualifyingscore.ScoreData, qualifyingRounds),
			}
			participants[participantID] = data
			result = append(result, *data)
		}

		if round.Valid && score.Valid {
			index := int(round.Int64) - 1

			if index < 0 || index >= len(data.Scores) {
				return nil, fmt.Errorf(
					"invalid qualifying round %d for participant %s",
					round.Int64,
					participantID,
				)
			}

			data.Scores[index].Score = int(score.Float64)
			data.Scores[index].Status = status
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate qualifying scores: %w", err)
	}

	for i := range result {
		result[i] = *participants[result[i].ParticipantID]
	}

	return result, nil
}

func (r *SQLiteQualifyingScoresRepository) FindScoresByParticipantID(
	ctx context.Context,
	tournamentID string,
	participantID string,
) ([]qualifyingscore.QualifyingScore, error) {
	query := `
        SELECT
            id,
            round,
            tournament_id,
            participant_id,
            score,
            created_at,
            updated_at
        FROM qualifying_scores
        WHERE tournament_id = ?
          AND participant_id = ?
        ORDER BY round ASC
    `

	rows, err := r.DB.QueryContext(ctx, query, tournamentID, participantID)
	if err != nil {
		return nil, fmt.Errorf("query qualifying scores by participant: %w", err)
	}
	defer rows.Close()

	scores := make([]qualifyingscore.QualifyingScore, 0)

	for rows.Next() {
		var score qualifyingscore.QualifyingScore

		if err := rows.Scan(
			&score.ID,
			&score.Round,
			&score.TournamentID,
			&score.ParticipantID,
			&score.Score,
			&score.CreatedAt,
			&score.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan qualifying score: %w", err)
		}

		scores = append(scores, score)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate qualifying scores: %w", err)
	}

	return scores, nil
}

func (r *SQLiteQualifyingScoresRepository) InitializeScores(
	ctx context.Context,
	tournamentID string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var qualifyingRounds int
	err = tx.QueryRowContext(ctx, `
        SELECT qualifying_rounds
        FROM tournaments
        WHERE id = ?
    `, tournamentID).Scan(&qualifyingRounds)
	if err != nil {
		return fmt.Errorf("get qualifying rounds: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
        SELECT id
        FROM participants
        WHERE tournament_id = ?
    `, tournamentID)
	if err != nil {
		return fmt.Errorf("get tournament participants: %w", err)
	}

	var participantIDs []string
	for rows.Next() {
		var participantID string
		if err := rows.Scan(&participantID); err != nil {
			rows.Close()
			return fmt.Errorf("scan participant ID: %w", err)
		}

		participantIDs = append(participantIDs, participantID)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate participants: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close participant rows: %w", err)
	}

	query := `
        INSERT INTO qualifying_scores (
            id,
            round,
            tournament_id,
            participant_id,
            score,
            created_at,
            updated_at
        )
        VALUES (?, ?, ?, ?, 0, ?, ?)
    `

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare qualifying score insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Unix()

	for _, participantID := range participantIDs {
		for round := 1; round <= qualifyingRounds; round++ {
			id, err := gonanoid.New(15)
			if err != nil {
				return err
			}
			_, err = stmt.ExecContext(
				ctx,
				id,
				round,
				tournamentID,
				participantID,
				now,
				now,
			)
			if err != nil {
				return fmt.Errorf(
					"initialize score for participant %s, round %d: %w",
					participantID,
					round,
					err,
				)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit qualifying scores: %w", err)
	}

	return nil
}

func (r *SQLiteQualifyingScoresRepository) UpdateScore(
	ctx context.Context,
	tournamentID string,
	participantID string,
	round int,
	score float64,
	status string,
) error {
	query := `
        UPDATE qualifying_scores
        SET score = ?, updated_at = ?, status = ?
        WHERE tournament_id = ?
          AND participant_id = ?
          AND round = ?
    `

	_, err := r.DB.ExecContext(
		ctx,
		query,
		score,
		time.Now().Unix(),
		status,
		tournamentID,
		participantID,
		round,
	)
	if err != nil {
		return fmt.Errorf("update qualifying score: %w", err)
	}

	return nil
}
