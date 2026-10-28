DROP TABLE IF EXISTS registered_car_photos;
DROP TABLE IF EXISTS registered_cars;
DROP TABLE IF EXISTS car_makes;

ALTER TABLE participants
DROP COLUMN driver_id;

DROP TABLE IF EXISTS drivers;