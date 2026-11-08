package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"judginginternal/internal/challonge"
	"judginginternal/internal/models/participant"
	"judginginternal/internal/models/tournament"
	"math/rand"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type BracketMatch struct {
	MatchID                string   `json:"matchID"`
	ParticipantName        *string  `json:"participantName"`
	Status                 string   `json:"status"`
	Round                  int      `json:"round"`
	WinnerID               *string  `json:"winnerId"`
	ChallongeParticipantID *string  `json:"challongeParticipantID"`
	ParticipantID          *string  `json:"participantID"`
	Score                  *float64 `json:"score"`
	BothRunFinished        bool     `json:"bothRunFinished"`
}

type TournamentRepository interface {
	// Creates tournament instance with it's configuration without having to have the
	// tournament initially set to qualifying state
	CreateTournament(ctx context.Context, req tournament.CreateRequest) (*tournament.Tournament, error)

	// InsertBulkParticipant adds participants (and their pending entrant rows) to an
	// existing tournament in a single transaction. The new participant IDs are
	// pushed onto the end of the tournament's qualifying_order. It returns the IDs
	// of the created participants, in the same order as the (trimmed, non-empty)
	// input.
	InsertBulkParticipant(ctx context.Context, tournamentID string, driverIDs []string) ([]string, error)

	// GetParticipants returns all participants registered to the given tournament,
	// ordered by registration time. It returns an empty slice if there are none.
	GetParticipants(ctx context.Context, tournamentID string) ([]participant.Participant, error)

	StartTournament(ctx context.Context, tournamentID string) error

	DeleteParticipant(ctx context.Context, tournamentID, participantID string) error
	GetTournaments() (tournaments []tournament.Tournament, err error)
	GetTournamentById(id string) (tournament.Tournament, error)
	DeleteTournamentById(ctx context.Context, id string) error
	UpdateTournament(ctx context.Context, id string, payload UpdateTournamentRequest) error
	UpdateTournamentState(ctx context.Context, id string, state string) error
}

type SQLiteTournamentRepository struct {
	DB              *sql.DB
	ChallongeAPIKey string
}

func (r *SQLiteTournamentRepository) CreateTournament(ctx context.Context, req tournament.CreateRequest) (*tournament.Tournament, error) {
	name := strings.TrimSpace(req.Name)

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.Size <= 0 || req.Size&(req.Size-1) != 0 {
		return nil, fmt.Errorf("size must be a positive power of two")
	}

	startingDate, err := time.Parse(time.RFC3339, req.StartingDate)
	if err != nil {
		startingDate, err = time.Parse("2006-01-02T15:04", req.StartingDate)
		if err != nil {
			return nil, fmt.Errorf("invalid starting date: %w", err)
		}
	}

	// create challonge tournament while we at it to save time later
	challongeRes, err := challonge.CreateTournament(
		ctx,
		r.ChallongeAPIKey,
		name,
		req.Size,
		startingDate,
	)
	if err != nil {
		return nil, err
	}
	challongeID := challongeRes.Data.ID
	challongeURL := challongeRes.Data.Attributes.URL

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	tournamentID, err := gonanoid.New(15)
	if err != nil {
		return nil, err
	}

	var criteriaConfig string
	err = tx.QueryRowContext(ctx, `
SELECT config from criteria_configs WHERE id = ?
`, req.CriteriaConfigID).Scan(&criteriaConfig)
	if err != nil {
		return nil, err
	}

	var qualifyingCriteriaConfig string
	err = tx.QueryRowContext(ctx, `
SELECT config from criteria_configs WHERE id = ?
`, req.QualifyingCriteriaConfigID).Scan(&qualifyingCriteriaConfig)
	if err != nil {
		return nil, err
	}

	// tournaments table
	var t tournament.Tournament

	err = tx.QueryRowContext(ctx, `
INSERT INTO tournaments (
    id, name, starting_date, track_id, qualifying_order,
    size, randomize_qualifying_order, status, created_at, updated_at, criteria_config, qualifying_criteria_config, qualifying_rounds
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, starting_date, track_id, qualifying_order,
          size, status, created_at, updated_at, criteria_config, qualifying_rounds
`,
		tournamentID,
		name,
		startingDate.Unix(),
		req.TrackID,
		"",
		req.Size,
		req.RandomizeQualifyingOrder,
		"pending", // set to pending since we dont want to start the qualifying immediately
		now,
		now,
		criteriaConfig,
		qualifyingCriteriaConfig,
		req.QualifyingRounds,
	).Scan(
		&t.ID,
		&t.Name,
		&t.StartingDate,
		&t.TrackID,
		&t.QualifyingOrder,
		&t.Size,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.CriteriaConfig,
		&t.QualifyingRounds,
	)
	if err != nil {
		return nil, err
	}

	bracketID, err := gonanoid.New(15)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO brackets (
	id, tournament_id, challonge_id, url, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, bracketID, tournamentID, challongeID, challongeURL, "not_started", now, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// remove challonge tournament if create tournament tx fails
		if err := challonge.DeleteTournament(ctx, r.ChallongeAPIKey, challongeID); err != nil {
			return nil, err
		}
		return nil, err
	}

	return &t, nil
}

func (r *SQLiteTournamentRepository) InsertBulkParticipant(
	ctx context.Context,
	tournamentID string,
	driverIDs []string,
) ([]string, error) {
	tournamentID = strings.TrimSpace(tournamentID)
	if tournamentID == "" {
		return nil, fmt.Errorf("tournament id is required")
	}

	ids := make([]string, 0, len(driverIDs))
	seen := make(map[string]struct{}, len(driverIDs))

	for _, id := range driverIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("at least one driver is required")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// ensure the tournament exists and get its current qualifying order.
	var existingOrder string
	err = tx.QueryRowContext(ctx, `
		SELECT qualifying_order
		FROM tournaments
		WHERE id = ?
	`, tournamentID).Scan(&existingOrder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tournament %s not found", tournamentID)
		}
		return nil, err
	}

	now := time.Now().Unix()

	// look up each driver's name and insert the linked participant.
	participantStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO participants (
			id, name, tournament_id, driver_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer participantStmt.Close()

	entrantStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO entrants (
			id, tournament_id, participant_id, status, score,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer entrantStmt.Close()

	participantIDs := make([]string, 0, len(ids))

	for _, driverID := range ids {
		var name string
		err := tx.QueryRowContext(ctx, `
			SELECT name
			FROM drivers
			WHERE id = ?
		`, driverID).Scan(&name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("driver %s not found", driverID)
			}
			return nil, err
		}

		participantID, err := gonanoid.New(15)
		if err != nil {
			return nil, err
		}

		_, err = participantStmt.ExecContext(
			ctx,
			participantID,
			name,
			tournamentID,
			driverID,
			now,
			now,
		)
		if err != nil {
			return nil, err
		}

		entrantID, err := gonanoid.New(15)
		if err != nil {
			return nil, err
		}

		_, err = entrantStmt.ExecContext(
			ctx,
			entrantID,
			tournamentID,
			participantID,
			"pending",
			0,
			now,
			now,
		)
		if err != nil {
			return nil, err
		}

		participantIDs = append(participantIDs, participantID)
	}

	// append the new participants to the qualifying order.
	newOrder := strings.Join(participantIDs, ", ")
	if strings.TrimSpace(existingOrder) != "" {
		newOrder = existingOrder + ", " + newOrder
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE tournaments
		SET qualifying_order = ?, updated_at = ?
		WHERE id = ?
	`, newOrder, now, tournamentID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return participantIDs, nil
}

func (r *SQLiteTournamentRepository) GetParticipants(
	ctx context.Context,
	tournamentID string,
) ([]participant.Participant, error) {
	tournamentID = strings.TrimSpace(tournamentID)
	if tournamentID == "" {
		return nil, fmt.Errorf("tournament id is required")
	}

	rows, err := r.DB.QueryContext(ctx, `
SELECT id, name, tournament_id, created_at, updated_at
FROM participants
WHERE tournament_id = ?
ORDER BY created_at ASC, rowid ASC
`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := make([]participant.Participant, 0)

	for rows.Next() {
		var p participant.Participant
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.TournamentID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *SQLiteTournamentRepository) StartTournament(
	ctx context.Context,
	tournamentID string,
) error {
	tournamentID = strings.TrimSpace(tournamentID)
	if tournamentID == "" {
		return fmt.Errorf("tournament id is required")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		status          string
		randomize       bool
		qualifyingOrder string
	)

	err = tx.QueryRowContext(ctx, `
SELECT status, randomize_qualifying_order, qualifying_order
FROM tournaments
WHERE id = ?
`, tournamentID).Scan(&status, &randomize, &qualifyingOrder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("tournament %s not found", tournamentID)
		}
		return err
	}

	if status != "pending" {
		return fmt.Errorf("tournament has already started (status: %s)", status)
	}

	// randomize qualifying order if enabled
	if randomize {
		ids := make([]string, 0)
		for _, id := range strings.Split(qualifyingOrder, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}

		rand.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		qualifyingOrder = strings.Join(ids, ", ")
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tournaments
SET status = ?, qualifying_order = ?, initial_qualifying_order = ?, updated_at = ?
WHERE id = ?
`, "qualifying", qualifyingOrder, qualifyingOrder, time.Now().Unix(), tournamentID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteTournamentRepository) DeleteParticipant(
	ctx context.Context,
	tournamentID string,
	participantID string,
) error {
	tournamentID = strings.TrimSpace(tournamentID)
	participantID = strings.TrimSpace(participantID)
	if tournamentID == "" || participantID == "" {
		return fmt.Errorf("tournament id and participant id are required")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// entrant row first, since it references the participant
	if _, err := tx.ExecContext(ctx, `
DELETE FROM entrants
WHERE tournament_id = ? AND participant_id = ?
`, tournamentID, participantID); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
DELETE FROM participants
WHERE id = ? AND tournament_id = ?
`, participantID, tournamentID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("participant %s not found in tournament %s", participantID, tournamentID)
	}

	// take the participant out of the qualifying order
	var existingOrder string
	if err := tx.QueryRowContext(ctx,
		`SELECT qualifying_order FROM tournaments WHERE id = ?`,
		tournamentID,
	).Scan(&existingOrder); err != nil {
		return err
	}

	remaining := make([]string, 0)
	for _, id := range strings.Split(existingOrder, ",") {
		if id = strings.TrimSpace(id); id != "" && id != participantID {
			remaining = append(remaining, id)
		}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE tournaments
SET qualifying_order = ?, updated_at = ?
WHERE id = ?
`, strings.Join(remaining, ", "), time.Now().Unix(), tournamentID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteTournamentRepository) setTournamentProgress(
	tournaments []tournament.Tournament,
) error {
	rows, err := r.DB.Query(`
		SELECT
			b.tournament_id,
			COUNT(DISTINCT n.match_id) AS total_matches,
			COUNT(DISTINCT CASE
				WHEN n.status = 'complete' THEN n.match_id
			END) AS completed_matches
		FROM brackets b
		JOIN nodes n ON n.bracket_id = b.id
		GROUP BY b.tournament_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type progress struct {
		total     int
		completed int
	}

	progressByTournament := make(map[string]progress)

	for rows.Next() {
		var tournamentID string
		var p progress

		if err := rows.Scan(
			&tournamentID,
			&p.total,
			&p.completed,
		); err != nil {
			return err
		}

		progressByTournament[tournamentID] = p
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for i := range tournaments {
		p := progressByTournament[tournaments[i].ID]

		if p.total == 0 {
			tournaments[i].Progress = 0
			continue
		}

		value := float64(p.completed) / float64(p.total) * 100

		if value < 0 {
			value = 0
		} else if value > 100 {
			value = 100
		}

		tournaments[i].Progress = value
	}

	return nil
}

func (r *SQLiteTournamentRepository) GetTournaments() ([]tournament.Tournament, error) {
	var tournaments []tournament.Tournament
	rows, err := r.DB.Query(`
SELECT 
	t1.id, t1.name, t1.starting_date, t2.name as track, t1.qualifying_order, t1.size, t1.status, t3.url,
	CASE WHEN t3.status = 'finalized' THEN 1 ELSE 0 END AS finalized
FROM tournaments t1
LEFT JOIN tracks t2
	ON t1.track_id = t2.id
LEFT JOIN brackets t3
	ON t1.id = t3.tournament_id
	`)
	if err != nil {
		return tournaments, err
	}
	defer rows.Close()

	for rows.Next() {
		var t tournament.Tournament

		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.StartingDate,
			&t.Track,
			&t.QualifyingOrder,
			&t.Size,
			&t.Status,
			&t.URL,
			&t.Finalized,
		); err != nil {
			return nil, err
		}

		tournaments = append(tournaments, t)
	}

	if err := rows.Err(); err != nil {
		return tournaments, err
	}

	if err := r.setTournamentProgress(tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func (r *SQLiteTournamentRepository) GetTournamentById(id string) (tournament.Tournament, error) {
	var t tournament.Tournament

	err := r.DB.QueryRow(`
SELECT
	t1.id,
	t1.name,
	t1.starting_date,
	t2.name AS track,
	t1.qualifying_order,
	t1.size,
	t1.status,
	t3.url,
	t1.qualifying_criteria_config,
	t1.criteria_config,
	t1.qualifying_rounds,
	t1.randomize_qualifying_order,
	t1.initial_qualifying_order,
	CASE WHEN t3.status = 'finalized' THEN 1 ELSE 0 END AS finalized
FROM tournaments t1
LEFT JOIN tracks t2
	ON t1.track_id = t2.id
LEFT JOIN brackets t3
	ON t1.id = t3.tournament_id
WHERE t1.id = ?
	`, id).Scan(
		&t.ID,
		&t.Name,
		&t.StartingDate,
		&t.Track,
		&t.QualifyingOrder,
		&t.Size,
		&t.Status,
		&t.URL,
		&t.QualifyingCriteriaConfig,
		&t.CriteriaConfig,
		&t.QualifyingRounds,
		&t.RandomizeQualifyingOrder,
		&t.InitialQualifyingOrder,
		&t.Finalized,
	)
	if err != nil {
		return tournament.Tournament{}, err
	}

	updated := []tournament.Tournament{t}
	if err := r.setTournamentProgress(updated); err != nil {
		return tournament.Tournament{}, err
	}

	return updated[0], nil
}

func (r *SQLiteTournamentRepository) DeleteTournamentById(
	ctx context.Context,
	id string,
) error {
	var challongeID string
	var status string

	err := r.DB.QueryRowContext(ctx, `
		SELECT challonge_id, status
		FROM brackets
		WHERE tournament_id = ?
		LIMIT 1
	`, id).Scan(&challongeID, &status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("bracket not found for this tournament")
		}
		return err
	}

	result, err := r.DB.ExecContext(ctx, `
		DELETE FROM tournaments
		WHERE id = ?
	`, id)
	if err != nil {
		return err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return sql.ErrNoRows
	}

	// only remove challonge's tournament entry that is not finalized yet
	if status != "finalized" {
		if err := challonge.DeleteTournament(ctx, r.ChallongeAPIKey, challongeID); err != nil {
			return err
		}
	}

	return nil
}

type UpdateTournamentRequest struct {
	Name            *string   `json:"name"`
	Size            *int      `json:"size"`
	StartingDate    *string   `json:"startingDate"`
	Track           *string   `json:"track"`
	QualifyingOrder *[]string `json:"qualifyingOrder"`
	Status          *string   `json:"status"`
}

func (r *SQLiteTournamentRepository) UpdateTournament(
	ctx context.Context,
	id string,
	payload UpdateTournamentRequest,
) error {
	var set []string
	var args []any

	if payload.Name != nil {
		set = append(set, "name = ?")
		args = append(args, *payload.Name)
	}

	if payload.Size != nil {
		set = append(set, "size = ?")
		args = append(args, *payload.Size)
	}

	if payload.StartingDate != nil {
		set = append(set, "starting_date = ?")
		args = append(args, *payload.StartingDate)
	}

	if payload.Track != nil {
		set = append(set, "track = ?")
		args = append(args, *payload.Track)
	}

	if payload.QualifyingOrder != nil {
		set = append(set, "qualifying_order = ?")
		args = append(args, strings.Join(*payload.QualifyingOrder, ", "))
	}

	if len(set) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE tournaments SET %s WHERE id = ?",
		strings.Join(set, ", "),
	)

	args = append(args, id)

	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update tournament: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("tournament %q not found", id)
	}

	return nil
}

func (r *SQLiteTournamentRepository) UpdateTournamentState(
	ctx context.Context,
	id, state string,
) error {
	var validStates = map[string]bool{
		"process_checkin":      true,
		"abort_checkin":        true,
		"start_group_stage":    true,
		"finalize_group_stage": true,
		"reset_group_stage":    true,
		"start":                true,
		"finalize":             true,
		"reset":                true,
		"open_predictions":     true,
	}

	if !validStates[state] {
		return fmt.Errorf("state %q is invalid", state)
	}

	var challongeURL string
	err := r.DB.QueryRowContext(ctx, `
		SELECT url
		FROM brackets
		WHERE tournament_id = ?
		LIMIT 1
	`, id).Scan(&challongeURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("bracket not found")
		}
		return err
	}

	if err := challonge.UpdateTournamentState(
		ctx,
		r.ChallongeAPIKey,
		challongeURL,
		state,
	); err != nil {
		return err
	}

	var status string

	switch state {
	case "start":
		status = "started"
	case "finalize":
		status = "finalized"
	case "reset":
		status = "not_started"
	}

	if status == "" {
		return nil
	}

	result, err := r.DB.ExecContext(ctx, `
		UPDATE brackets
		SET status = ?
		WHERE tournament_id = ?
	`, status, id)
	if err != nil {
		return err
	}

	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if updated == 0 {
		return fmt.Errorf("cannot update bracket status to %s", status)
	}

	return nil
}
