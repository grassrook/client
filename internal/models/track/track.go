package track

type Track struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type CreateTrackReq struct {
	Name string `json:"name"`
}

type UpdateTrackReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
