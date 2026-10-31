package tournament

type Tournament struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	StartingDate             string  `json:"starting_date"`
	TrackID                  string  `json:"track_id"`
	Track                    *string `json:"track"`
	QualifyingOrder          string  `json:"qualifying_order"`
	Status                   string  `json:"status"`
	URL                      string  `json:"url"`
	Size                     int     `json:"size"`
	Progress                 float64 `json:"progress"`
	CriteriaConfig           string  `json:"criteriaConfig"`
	Finalized                bool    `json:"finalized"`
	QualifyingRounds         int     `json:"qualifyingRounds"`
	RandomizeQualifyingOrder bool    `json:"randomizeQualifyingOrder"`
	InitialQualifyingOrder   *string `json:"initialQualifyingOrder"`
	CreatedAt                int64   `json:"createdAt"`
	UpdatedAt                int64   `json:"updatedAt"`
}

type CreateRequest struct {
	Name                     string `json:"name"`
	Size                     int    `json:"size"`
	StartingDate             string `json:"startingDate"`
	TrackID                  string `json:"trackID"`
	CriteriaConfigID         int    `json:"criteriaConfigID"`
	QualifyingRounds         int    `json:"qualifyingRounds"`
	RandomizeQualifyingOrder bool   `json:"randomizeQualifyingOrder"`
}
