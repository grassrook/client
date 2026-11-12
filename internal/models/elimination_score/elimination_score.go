package eliminationscore

type EliminationScore struct {
	ID           string `json:"id"`
	Round        int    `json:"round"`
	TournamentID string `json:"tournamentId"`
	MatchID      string `json:"matchId"`
	Status       string `json:"status"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

type Result struct {
	Lead  *float64 `json:"lead,omitempty"`
	Chase *float64 `json:"chase,omitempty"`
}

type Results map[string]Result

type ScoreData struct {
	ParticipantID     string  `json:"participantID"`
	LeadParticipantID string  `json:"leadParticipantID"`
	Score             float64 `json:"score"`
	Status            string  `json:"status"`
	Results           Results `json:"results"`
	Round             int     `json:"round"`
}

type Data struct {
	Scores []ScoreData `json:"scores"` // each indices represent round
}

type InitializeScoreParam struct {
	TournamentID string   `json:"tournamentID"`
	MatchID      string   `json:"matchId"`
	Participants []string `json:"participants"` // first index goes lead first
}

type UpdateEliminationScoreParams struct {
	MatchID           string  `json:"matchId"`
	LeadParticipantID string  `json:"leadParticipantID"`
	Results           Results `json:"results"`
	Score             float64 `json:"score"`
	ParticipantID     string  `json:"participantID"`
}

type UpdateEliminationScoreByRoundParams struct {
	MatchID       string  `json:"matchId"`
	Round         int     `json:"round"`
	Results       Results `json:"results"`
	Score         float64 `json:"score"`
	ParticipantID string  `json:"participantID"`
}
