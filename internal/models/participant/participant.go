package participant

type Participant struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	TournamentID           string  `json:"tournament_id"`
	ChallongeParticipantID *string `json:"ChallongeParticipantID"`
	DriverID               string  `json:"driverID"`
	RegisteredCarID        string  `json:"registeredCarID"`
	CreatedAt              int64   `json:"created_at"`
	UpdatedAt              int64   `json:"updated_at"`
}

type ReplaceParticipantResult struct {
	From Participant `json:"from"`
	To   Participant `json:"to"`
}
