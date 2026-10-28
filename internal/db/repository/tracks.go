package repository

import (
	"context"
	"database/sql"
	"judginginternal/internal/models/track"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type TracksRepository interface {
	GetAll(ctx context.Context) ([]track.Track, error)
	Create(ctx context.Context, t track.CreateTrackReq) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, t track.UpdateTrackReq) error
}

type SQLiteTracksRepository struct {
	DB *sql.DB
}

func NewSQLiteTracksRepository(db *sql.DB) *SQLiteTracksRepository {
	return &SQLiteTracksRepository{
		DB: db,
	}
}

func (r *SQLiteTracksRepository) Create(ctx context.Context, t track.CreateTrackReq) error {
	id, err := gonanoid.New()
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO tracks (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, id, t.Name, now, now)

	return err
}

func (r *SQLiteTracksRepository) GetAll(ctx context.Context) ([]track.Track, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM tracks
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]track.Track, 0)

	for rows.Next() {
		var t track.Track

		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tracks = append(tracks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *SQLiteTracksRepository) Delete(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `
		DELETE FROM tracks
		WHERE id = ?
	`, id)

	return err
}

func (r *SQLiteTracksRepository) Update(ctx context.Context, t track.UpdateTrackReq) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE tracks
		SET name = ?, updated_at = ?
		WHERE id = ?
	`, t.Name, time.Now().Unix(), t.ID)

	return err
}
