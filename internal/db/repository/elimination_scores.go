package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	eliminationscore "judginginternal/internal/models/elimination_score"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type EliminationScoresRepository interface {
	InitializeScores(ctx context.Context, param eliminationscore.InitializeScoreParam) ([]eliminationscore.ScoreData, error)
	UpdateScore(ctx context.Context, params eliminationscore.UpdateEliminationScoreParams) (float64, error)
	UpdateScoreByRound(ctx context.Context, params eliminationscore.UpdateEliminationScoreByRoundParams) (float64, error)
	ResetScores(ctx context.Context, matchID string) error
}

type SQLiteEliminationScoresRepository struct {
	DB *sql.DB
}

func (r *SQLiteEliminationScoresRepository) InitializeScores(
	ctx context.Context,
	param eliminationscore.InitializeScoreParam,
) ([]eliminationscore.ScoreData, error) {
	if len(param.Participants) != 2 {
		return nil, fmt.Errorf("expected exactly 2 participants")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM scores
			WHERE tournament_id = ?
				AND match_id = ?
				AND kind = 'elimination'
		)
	`, param.TournamentID, param.MatchID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check existing elimination scores: %w", err)
	}

	if !exists {
		var leadParticipantID string

		err = tx.QueryRowContext(ctx, `
			SELECT participant_id
			FROM active_battles
			WHERE match_id = ?
				AND lead_status = 1
			LIMIT 1
		`, param.MatchID).Scan(&leadParticipantID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("no leading participant found for match %s", param.MatchID)
			}
			return nil, fmt.Errorf("find leading participant: %w", err)
		}

		if leadParticipantID != param.Participants[0] &&
			leadParticipantID != param.Participants[1] {
			return nil, fmt.Errorf("leading participant does not belong to this match")
		}

		chasingParticipantID := param.Participants[0]
		if chasingParticipantID == leadParticipantID {
			chasingParticipantID = param.Participants[1]
		}

		now := time.Now().Unix()

		for round := 1; round <= 2; round++ {
			roundLeaderID := leadParticipantID
			if round == 2 {
				roundLeaderID = chasingParticipantID
			}

			for _, participantID := range param.Participants {
				id, err := gonanoid.New(15)
				if err != nil {
					return nil, fmt.Errorf("generate elimination score ID: %w", err)
				}

				_, err = tx.ExecContext(ctx, `
					INSERT INTO scores (
						id,
						tournament_id,
						participant_id,
						match_id,
						round,
						kind,
						results,
						score,
						status,
						lead_participant_id,
						created_at,
						updated_at
					) VALUES (?, ?, ?, ?, ?, 'elimination', '{}', 0, 'pending', ?, ?, ?)
				`,
					id,
					param.TournamentID,
					participantID,
					param.MatchID,
					round,
					roundLeaderID,
					now,
					now,
				)
				if err != nil {
					return nil, fmt.Errorf("insert elimination score: %w", err)
				}
			}
		}
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			participant_id,
			lead_participant_id,
			score,
			status,
			results,
			round
		FROM scores
		WHERE tournament_id = ?
			AND match_id = ?
			AND kind = 'elimination'
		ORDER BY round, participant_id
	`, param.TournamentID, param.MatchID)
	if err != nil {
		return nil, fmt.Errorf("query elimination scores: %w", err)
	}

	scores := make([]eliminationscore.ScoreData, 0, 4)

	for rows.Next() {
		var s eliminationscore.ScoreData
		var results []byte

		if err := rows.Scan(
			&s.ParticipantID,
			&s.LeadParticipantID,
			&s.Score,
			&s.Status,
			&results,
			&s.Round,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan elimination score: %w", err)
		}

		if err := json.Unmarshal(results, &s.Results); err != nil {
			rows.Close()
			return nil, fmt.Errorf("unmarshal elimination results: %w", err)
		}

		scores = append(scores, s)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate elimination scores: %w", err)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close elimination score rows: %w", err)
	}

	if len(scores) != 4 {
		return nil, fmt.Errorf("expected 4 elimination score rows, got %d", len(scores))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return scores, nil
}

func (r *SQLiteEliminationScoresRepository) UpdateScore(ctx context.Context, params eliminationscore.UpdateEliminationScoreParams) (float64, error) {
	results, err := json.Marshal(params.Results)
	if err != nil {
		return 0, fmt.Errorf("marshal elimination score results: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `
		UPDATE scores
		SET
			results = ?,
			updated_at = ?,
			score = ?,
			status = 'scored'
		WHERE lead_participant_id = ?
			AND match_id = ?
			AND participant_id = ?
			AND kind = 'elimination'
	`,
		string(results),
		time.Now().Unix(),
		params.Score,
		params.LeadParticipantID,
		params.MatchID,
		params.ParticipantID,
	)
	if err != nil {
		return 0, fmt.Errorf("update elimination score: %w", err)
	}

	return params.Score, nil
}

func (r *SQLiteEliminationScoresRepository) UpdateScoreByRound(
	ctx context.Context,
	params eliminationscore.UpdateEliminationScoreByRoundParams,
) (float64, error) {
	results, err := json.Marshal(params.Results)
	if err != nil {
		return 0, fmt.Errorf("marshal elimination score results: %w", err)
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE scores
		SET
			results = ?,
			updated_at = ?,
			score = ?,
			status = 'scored'
		WHERE match_id = ?
			AND participant_id = ?
			AND round = ?
			AND kind = 'elimination'
	`,
		string(results),
		time.Now().Unix(),
		params.Score,
		params.MatchID,
		params.ParticipantID,
		params.Round,
	)
	if err != nil {
		return 0, fmt.Errorf("update elimination score by round: %w", err)
	}

	var total float64

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(score), 0)
		FROM scores
		WHERE match_id = ?
			AND participant_id = ?
			AND kind = 'elimination'
			AND round IN (1, 2)
	`, params.MatchID, params.ParticipantID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("calculate total elimination score: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return total, nil
}

func (r *SQLiteEliminationScoresRepository) ResetScores(
	ctx context.Context,
	matchID string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var leadParticipantID string

	err = tx.QueryRowContext(ctx, `
		SELECT participant_id
		FROM active_battles
		WHERE match_id = ?
			AND lead_status = 1
		LIMIT 1
	`, matchID).Scan(&leadParticipantID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no leading participant found for match %s", matchID)
		}
		return fmt.Errorf("find leading participant: %w", err)
	}

	var chasingParticipantID string

	err = tx.QueryRowContext(ctx, `
		SELECT participant_id
		FROM active_battles
		WHERE match_id = ?
			AND lead_status = 0
		LIMIT 1
	`, matchID).Scan(&chasingParticipantID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no chasing participant found for match %s", matchID)
		}
		return fmt.Errorf("find chasing participant: %w", err)
	}

	now := time.Now().Unix()

	leaders := [2]string{
		leadParticipantID,
		chasingParticipantID,
	}

	for round := 1; round <= 2; round++ {
		_, err := tx.ExecContext(ctx, `
			UPDATE scores
			SET
				lead_participant_id = ?,
				results = '{}',
				status = 'pending',
				score = 0,
				updated_at = ?
			WHERE match_id = ?
				AND round = ?
				AND kind = 'elimination'
		`,
			leaders[round-1],
			now,
			matchID,
			round,
		)
		if err != nil {
			return fmt.Errorf("reset elimination scores for round %d: %w", round, err)
		}
	}

	return tx.Commit()
}
