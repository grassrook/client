package qualifyingscore

type QualifyingScore struct {
	ID            string  `json:"id"`
	Round         int     `json:"round"`
	TournamentID  string  `json:"tournamentId"`
	ParticipantID string  `json:"participantId"`
	Score         float64 `json:"score"`
	Status        string  `json:"status"`
	CreatedAt     int64   `json:"createdAt"`
	UpdatedAt     int64   `json:"updatedAt"`
}

type ScoreData struct {
	Score  float64 `json:"score"`
	Status string  `json:"status"`
}

type Data struct {
	ParticipantID string      `json:"participantId"`
	Scores        []ScoreData `json:"scores"` // each indices represent round
}
