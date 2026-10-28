CREATE TABLE drivers (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE car_makes (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE registered_cars (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    car_make_id TEXT NOT NULL REFERENCES car_makes(id),
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE registered_car_photos (
    id TEXT PRIMARY KEY NOT NULL,
    registered_car_id TEXT NOT NULL
        REFERENCES registered_cars(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    created_at INTEGER NOT NULL
);