package repository

import (
	"context"
	"database/sql"
	"time"
)

type ParticipantsRepo interface {
	UpdateRegisteredCar(
		ctx context.Context,
		participantID string,
		registeredCarID string,
	) error
}

type SQLiteParticipantsRepository struct {
	DB *sql.DB
}

func (r *SQLiteParticipantsRepository) UpdateRegisteredCar(
	ctx context.Context,
	participantID string,
	registeredCarID string,
) error {
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
