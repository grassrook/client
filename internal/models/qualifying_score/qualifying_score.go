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

type Result struct {
	Lead  *float64 `json:"lead,omitempty"`
	Chase *float64 `json:"chase,omitempty"`
}

type Results map[string]Result

type ScoreData struct {
	Score   float64 `json:"score"`
	Status  string  `json:"status"`
	Results Results `json:"results"`
}

type Data struct {
	ParticipantID string      `json:"participantId"`
	Scores        []ScoreData `json:"scores"` // each indices represent round
}

type UpdateQualifyingScoreParams struct {
	TournamentID  string
	ParticipantID string
	Round         int
	Score         float64
	Status        string
	Results       Results
}
