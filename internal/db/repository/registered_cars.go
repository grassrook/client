package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"judginginternal/internal/models/car"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrRegisteredCarNotFound = errors.New("registered car not found")

type RegisteredCarsRepository interface {
	Create(ctx context.Context, input car.CreateRegisteredCar) error
	GetByID(ctx context.Context, id string) (car.RegisteredCar, error)
	List(ctx context.Context) ([]car.RegisteredCar, error)
	ListCarMakes(ctx context.Context) ([]car.CarMake, error)
	ListByDriverID(ctx context.Context, driverID string) ([]car.RegisteredCar, error)
	Update(ctx context.Context, id string, c car.UpdateRegisteredCar) error
	Delete(ctx context.Context, id string) error
}

type SQLiteRegisteredCarsRepository struct {
	DB        *sql.DB
	MediaPath string
}

func imageExtension(mimeType string) (string, error) {
	switch mimeType {
	case "data:image/jpeg;base64":
		return ".jpg", nil
	case "data:image/png;base64":
		return ".png", nil
	case "data:image/webp;base64":
		return ".webp", nil
	default:
		return "", errors.New("unsupported image format")
	}
}

func (r *SQLiteRegisteredCarsRepository) Create(
	ctx context.Context,
	input car.CreateRegisteredCar,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	carID, err := gonanoid.New(15)
	if err != nil {
		return err
	}

	mediaDir := filepath.Join(r.MediaPath, input.DriverID)
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return err
	}

	var savedFiles []string

	committed := false
	defer func() {
		if !committed {
			for _, path := range savedFiles {
				_ = os.Remove(path)
			}
		}
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO registered_cars (
			id, name, description, car_make_id,
			driver_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		carID,
		input.Name,
		input.Description,
		input.CarMakeID,
		input.DriverID,
		now,
		now,
	)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO registered_car_photos (
			id, registered_car_id, path, created_at
		)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, photo := range input.Photos {
		mimeType, encoded, ok := strings.Cut(photo, ",")
		if !ok || !strings.HasPrefix(mimeType, "data:image/") ||
			!strings.HasSuffix(mimeType, ";base64") {
			return errors.New("invalid image data URL")
		}

		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return err
		}

		ext, err := imageExtension(mimeType)
		if err != nil {
			return err
		}

		filenameID, err := gonanoid.New(15)
		if err != nil {
			return err
		}
		filename := filenameID + ext
		destination := filepath.Join(mediaDir, filename)

		if err := os.WriteFile(destination, data, 0644); err != nil {
			return err
		}

		savedFiles = append(savedFiles, destination)

		relativePath, err := filepath.Rel(r.MediaPath, destination)
		if err != nil {
			return err
		}

		relativePath = filepath.ToSlash(relativePath)

		if _, err := stmt.ExecContext(
			ctx,
			filenameID,
			carID,
			relativePath,
			now,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true
	return nil
}

func (r *SQLiteRegisteredCarsRepository) GetByID(
	ctx context.Context,
	id string,
) (car.RegisteredCar, error) {
	var c car.RegisteredCar

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, name, description, car_make_id, driver_id,
		       created_at, updated_at
		FROM registered_cars
		WHERE id = ?
	`, id).Scan(
		&c.ID,
		&c.Name,
		&c.Description,
		&c.CarMakeID,
		&c.DriverID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return car.RegisteredCar{}, ErrRegisteredCarNotFound
	}
	if err != nil {
		return car.RegisteredCar{}, err
	}

	cars := []car.RegisteredCar{c}
	if err := r.loadPhotos(ctx, cars); err != nil {
		return car.RegisteredCar{}, err
	}

	return cars[0], nil
}

func (r *SQLiteRegisteredCarsRepository) List(
	ctx context.Context,
) ([]car.RegisteredCar, error) {
	return r.list(ctx, `
		SELECT id, name, description, car_make_id, driver_id,
		       created_at, updated_at
		FROM registered_cars
		ORDER BY created_at DESC
	`)
}

func (r *SQLiteRegisteredCarsRepository) ListCarMakes(
	ctx context.Context,
) ([]car.CarMake, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name
		FROM car_makes
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	makes := make([]car.CarMake, 0)

	for rows.Next() {
		var make car.CarMake

		if err := rows.Scan(&make.ID, &make.Name); err != nil {
			return nil, err
		}

		makes = append(makes, make)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return makes, nil
}

func (r *SQLiteRegisteredCarsRepository) ListByDriverID(
	ctx context.Context,
	driverID string,
) ([]car.RegisteredCar, error) {
	return r.list(ctx, `
		SELECT id, name, description, car_make_id, driver_id,
		       created_at, updated_at
		FROM registered_cars
		WHERE driver_id = ?
		ORDER BY created_at DESC
	`, driverID)
}

func (r *SQLiteRegisteredCarsRepository) list(
	ctx context.Context,
	query string,
	args ...any,
) ([]car.RegisteredCar, error) {
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := make([]car.RegisteredCar, 0)

	for rows.Next() {
		var c car.RegisteredCar

		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Description,
			&c.CarMakeID,
			&c.DriverID,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		cars = append(cars, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.loadPhotos(ctx, cars); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *SQLiteRegisteredCarsRepository) loadPhotos(
	ctx context.Context,
	cars []car.RegisteredCar,
) error {
	if len(cars) == 0 {
		return nil
	}

	indices := make(map[string]int, len(cars))
	for i := range cars {
		cars[i].Photos = make([]string, 0)
		indices[cars[i].ID] = i
	}

	ids := make([]any, len(cars))
	for i, c := range cars {
		ids[i] = c.ID
	}

	query := `
		SELECT registered_car_id, path
		FROM registered_car_photos
		WHERE registered_car_id IN (` +
		strings.TrimSuffix(strings.Repeat("?,", len(cars)), ",") + `)
		ORDER BY created_at ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var carID, path string

		if err := rows.Scan(&carID, &path); err != nil {
			return err
		}

		cars[indices[carID]].Photos = append(
			cars[indices[carID]].Photos,
			path,
		)
	}

	return rows.Err()
}

func (r *SQLiteRegisteredCarsRepository) Update(
	ctx context.Context,
	id string,
	input car.UpdateRegisteredCar,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRowContext(ctx, `
		SELECT 1 FROM registered_cars WHERE id = ?
	`, id).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrRegisteredCarNotFound
	}
	if err != nil {
		return err
	}

	var (
		sets []string
		args []any
	)

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}

	if input.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *input.Description)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = ?")
		args = append(args, time.Now().Unix())
		args = append(args, id)

		query := `
			UPDATE registered_cars
			SET ` + strings.Join(sets, ", ") + `
			WHERE id = ?
		`

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	if input.Photos != nil {
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM registered_car_photos
			WHERE registered_car_id = ?
		`, id); err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO registered_car_photos (
				id, registered_car_id, path, created_at
			)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		now := time.Now().Unix()

		for _, path := range *input.Photos {
			photoID := uuid.NewString()

			if _, err := stmt.ExecContext(
				ctx,
				photoID,
				id,
				path,
				now,
			); err != nil {
				return err
			}
		}

		if len(sets) == 0 {
			if _, err := tx.ExecContext(ctx, `
				UPDATE registered_cars
				SET updated_at = ?
				WHERE id = ?
			`, now, id); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteRegisteredCarsRepository) Delete(
	ctx context.Context,
	id string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRowContext(ctx, `
		SELECT 1
		FROM registered_cars
		WHERE id = ?
	`, id).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrRegisteredCarNotFound
	}
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT path
		FROM registered_car_photos
		WHERE registered_car_id = ?
	`, id)
	if err != nil {
		return err
	}

	var photos []string

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			rows.Close()
			return err
		}

		photos = append(photos, path)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	if err := rows.Close(); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM registered_car_photos
		WHERE registered_car_id = ?
	`, id)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM registered_cars
		WHERE id = ?
	`, id)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, photo := range photos {
		// Prevent paths from escaping MediaPath.
		fullPath := filepath.Join(r.MediaPath, filepath.Clean(photo))

		rel, err := filepath.Rel(r.MediaPath, fullPath)
		if err != nil || rel == ".." ||
			strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}

		if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Database deletion already succeeded.
			// Log this error if a logger is available.
			continue
		}
	}

	return nil
}
