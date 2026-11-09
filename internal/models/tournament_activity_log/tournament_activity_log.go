package tournamentactivitylog

type LogKind = string

const (
	TOURNAMENT_INIT      LogKind = "TOURNAMENT_INIT"
	TOURNAMENT_START     LogKind = "TOURNAMENT_START"
	QUALIFYING_FINISHED  LogKind = "QUALIFYING_FINISHED"
	WON_BATTLE           LogKind = "WON_BATTLE"
	OMT                  LogKind = "OMT"
	QUALIFYING_SCORED    LogKind = "QUALIFYING_SCORED"
	NOT_QUALIFIED        LogKind = "NOT_QUALIFIED"
	QUALIFYING_RESET     LogKind = "QUALIFYING_RESET"
	TOURNAMENT_FINISHED  LogKind = "TOURNAMENT_FINISHED"
	REPLACE_PARTICIPANTS LogKind = "REPLACE_PARTICIPANTS"
)

type TournamentActivityLog struct {
	ID           string  `json:"id"`
	TournamentID string  `json:"tournament_id"`
	Content      string  `json:"content"`
	LogType      LogKind `json:"logType"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
}

type CreateTournamentActivityLogInput struct {
	TournamentID string  `json:"tournament_id"`
	Content      string  `json:"content"`
	LogType      LogKind `json:"logType"`
}
