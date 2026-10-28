package repository

import (
	"context"
	"database/sql"
	"time"

	"judginginternal/internal/models/driver"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type DriversRepository interface {
	GetAll(ctx context.Context) ([]driver.Driver, error)
	GetByID(ctx context.Context, id string) (driver.Driver, error)
	Create(ctx context.Context, req driver.CreateDriverReq) error
	Update(ctx context.Context, req driver.UpdateDriverReq) error
	Delete(ctx context.Context, id string) error
}

type SQLiteDriversRepository struct {
	DB *sql.DB
}

func (r *SQLiteDriversRepository) GetAll(
	ctx context.Context,
) ([]driver.Driver, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM drivers
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	drivers := make([]driver.Driver, 0)

	for rows.Next() {
		var d driver.Driver

		if err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.Description,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, err
		}

		drivers = append(drivers, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drivers, nil
}

func (r *SQLiteDriversRepository) GetByID(
	ctx context.Context,
	id string,
) (driver.Driver, error) {
	var d driver.Driver

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM drivers
		WHERE id = ?
	`, id).Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return driver.Driver{}, err
	}

	return d, nil
}

func (r *SQLiteDriversRepository) Create(
	ctx context.Context,
	req driver.CreateDriverReq,
) error {
	id, err := gonanoid.New()
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO drivers (
			id, name, description, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		id,
		req.Name,
		req.Description,
		now,
		now,
	)

	return err
}

func (r *SQLiteDriversRepository) Update(
	ctx context.Context,
	req driver.UpdateDriverReq,
) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE drivers
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`,
		req.Name,
		req.Description,
		time.Now().Unix(),
		req.ID,
	)

	return err
}

func (r *SQLiteDriversRepository) Delete(
	ctx context.Context,
	id string,
) error {
	_, err := r.DB.ExecContext(ctx, `
		DELETE FROM drivers
		WHERE id = ?
	`, id)

	return err
}
