package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	UpdateScore(ctx context.Context, params qualifyingscore.UpdateQualifyingScoreParams) error
}

type SQLiteQualifyingScoresRepository struct {
	DB *sql.DB
}

func (r *SQLiteQualifyingScoresRepository) FindAllScores(
	ctx context.Context,
	tournamentID string,
) ([]qualifyingscore.Data, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			p.id,
			s.round,
			s.score,
			s.results,
			s.status
		FROM participants p
		LEFT JOIN scores s
			ON s.participant_id = p.id
			AND s.kind = 'qualifying'
		WHERE p.tournament_id = ?
		ORDER BY p.id, s.round
	`, tournamentID)
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
		var status sql.NullString
		var results []byte

		if err := rows.Scan(&participantID, &round, &score, &results, &status); err != nil {
			return nil, fmt.Errorf("scan qualifying score: %w", err)
		}

		data, exists := participants[participantID]
		if !exists {
			data = &qualifyingscore.Data{
				ParticipantID: participantID,
				Scores:        make([]qualifyingscore.ScoreData, 0),
			}
			participants[participantID] = data
			result = append(result, *data)
		}

		if round.Valid {
			index := int(round.Int64) - 1

			for len(data.Scores) <= index {
				data.Scores = append(data.Scores, qualifyingscore.ScoreData{})
			}

			var parsedResults qualifyingscore.Results

			if len(results) > 0 {
				if err := json.Unmarshal(results, &parsedResults); err != nil {
					return nil, fmt.Errorf("unmarshal qualifying results: %w", err)
				}
			}

			data.Scores[index].Score = score.Float64
			data.Scores[index].Status = status.String
			data.Scores[index].Results = parsedResults
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
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			s.id,
			s.round,
			p.tournament_id,
			s.participant_id,
			s.score,
			s.created_at,
			s.updated_at
		FROM scores s
		JOIN participants p ON p.id = s.participant_id
		WHERE p.tournament_id = ?
			AND s.participant_id = ?
			AND s.kind = 'qualifying'
		ORDER BY s.round ASC
	`, tournamentID, participantID)
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
		INSERT INTO scores (
			id,
			tournament_id,
			participant_id,
			round,
			kind,
			score,
			status,
			results,
			created_at,
			updated_at
		)
		VALUES (?, ?,  ?, ?, 'qualifying', 0, 'pending', '{}', ?, ?)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare qualifying score insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Unix()

	for _, participantID := range participantIDs {
		var maxRound int

		err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(round), 0)
			FROM scores
			WHERE participant_id = ?
				AND kind = 'qualifying'
		`, participantID).Scan(&maxRound)
		if err != nil {
			return fmt.Errorf("get existing qualifying rounds: %w", err)
		}

		var qualifyingRounds int

		err = tx.QueryRowContext(ctx, `
			SELECT qualifying_rounds
			FROM tournaments
			WHERE id = ?
		`, tournamentID).Scan(&qualifyingRounds)
		if err != nil {
			return fmt.Errorf("get qualifying rounds: %w", err)
		}

		for round := maxRound + 1; round <= qualifyingRounds; round++ {
			id, err := gonanoid.New(15)
			if err != nil {
				return fmt.Errorf("generate score ID: %w", err)
			}

			_, err = stmt.ExecContext(
				ctx,
				id,
				tournamentID,
				participantID,
				round,
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

func (r *SQLiteQualifyingScoresRepository) UpdateScore(ctx context.Context, params qualifyingscore.UpdateQualifyingScoreParams) error {
	results, err := json.Marshal(params.Results)
	if err != nil {
		return fmt.Errorf("marshal qualifying score results: %w", err)
	}

	_, err = r.DB.ExecContext(ctx, `
		UPDATE scores
		SET
			score = ?,
			status = ?,
			results = ?,
			updated_at = ?
		WHERE participant_id = ?
			AND round = ?
			AND kind = 'qualifying'
			AND participant_id IN (
				SELECT id
				FROM participants
				WHERE tournament_id = ?
			)
	`,
		params.Score,
		params.Status,
		string(results),
		time.Now().Unix(),
		params.ParticipantID,
		params.Round,
		params.TournamentID,
	)
	if err != nil {
		return fmt.Errorf("update qualifying score: %w", err)
	}

	return nil
}
