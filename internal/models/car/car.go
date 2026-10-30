package car

type RegisteredCar struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CarMakeID   string   `json:"carMakeID"`
	DriverID    string   `json:"driverID"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
	Photos      []string `json:"photos"`
}

type RegisteredCarPhoto struct {
	ID              string `json:"id"`
	RegisteredCarID string `json:"registeredCarID"`
	Path            string `json:"path"`
	CreatedAt       int64  `json:"createdAt"`
}

type CreateRegisteredCar struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CarMakeID   string   `json:"carMakeID"`
	DriverID    string   `json:"driverID"`
	Photos      []string `json:"photos"`
}

type UpdateRegisteredCar struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Photos      *[]string `json:"photos,omitempty"`
}

type CarMake struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
