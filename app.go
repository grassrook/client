package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"judginginternal/internal/config"
	"judginginternal/internal/db"
	"judginginternal/internal/db/repository"
	"judginginternal/internal/logger"
	"judginginternal/internal/models/car"
	"judginginternal/internal/models/driver"
	eliminationscore "judginginternal/internal/models/elimination_score"
	"judginginternal/internal/models/entrant"
	"judginginternal/internal/models/participant"
	qualifyingscore "judginginternal/internal/models/qualifying_score"
	"judginginternal/internal/models/tournament"
	tournamentactivitylog "judginginternal/internal/models/tournament_activity_log"
	"judginginternal/internal/models/track"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/amoghe/distillog"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var client *sql.DB = nil

// App struct
type App struct {
	ctx    context.Context
	cfg    *config.Config
	logger distillog.Logger

	tournamentRepo             repository.TournamentRepository
	tracksRepo                 repository.TracksRepository
	driversRepo                repository.DriversRepository
	matchesRepo                repository.MatchesRepository
	bracketRepo                repository.BracketRepository
	participantsRepo           repository.ParticipantsRepo
	entrantsRepo               repository.EntrantsRepository
	criteriaConfigRepo         repository.CriteriaConfigRepository
	qualifyingScoresRepo       repository.QualfyingScoresRepository
	eliminationScoresRepo      repository.EliminationScoresRepository
	tournamentActivityLogsRepo repository.TournamentActivityLogsRepository
	registeredCarsRepo         repository.RegisteredCarsRepository
}

// NewApp creates a new App application struct
func NewApp(cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		logger: logger.New(cfg.Workspace),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	client = db.Client(path.Join(a.cfg.Workspace, "sqlite.db"))

	a.tournamentRepo = &repository.SQLiteTournamentRepository{
		DB:              client,
		ChallongeAPIKey: a.cfg.ChallongeApiKey,
	}

	a.tracksRepo = &repository.SQLiteTracksRepository{
		DB: client,
	}

	a.driversRepo = &repository.SQLiteDriversRepository{
		DB: client,
	}

	a.matchesRepo = &repository.SQLiteMatchesRepository{
		DB:              client,
		ChallongeAPIKey: a.cfg.ChallongeApiKey,
	}

	a.bracketRepo = &repository.SQLiteBracketRepository{
		DB:              client,
		ChallongeAPIKey: a.cfg.ChallongeApiKey,
	}

	a.participantsRepo = &repository.SQLiteParticipantsRepository{
		DB:              client,
		ChallongeAPIKey: a.cfg.ChallongeApiKey,
	}

	a.entrantsRepo = &repository.SQLiteEntrantsRepository{
		DB:              client,
		ChallongeAPIKey: a.cfg.ChallongeApiKey,
	}

	a.criteriaConfigRepo = &repository.SQLiteCriteriaConfigRepository{
		DB: client,
	}

	a.qualifyingScoresRepo = &repository.SQLiteQualifyingScoresRepository{
		DB: client,
	}

	a.eliminationScoresRepo = &repository.SQLiteEliminationScoresRepository{
		DB: client,
	}

	a.tournamentActivityLogsRepo = &repository.SQLiteTournamentActivityLogsRepository{
		DB: client,
	}

	a.registeredCarsRepo = &repository.SQLiteRegisteredCarsRepository{
		DB:        client,
		MediaPath: path.Join(a.cfg.Workspace, "media"),
	}

	runtime.LogInfo(a.ctx, "Database connection established and repositories initialized")
}

func (a *App) logError(operation string, err error) {
	if err == nil {
		return
	}

	message := fmt.Sprintf("%s: %v", operation, err)

	runtime.LogError(a.ctx, message)
	a.logger.Errorf("%s", message)
}

func (a *App) GetCarPhoto(photoPath string) (string, error) {
	if photoPath == "" {
		err := fmt.Errorf("photo path is empty")
		a.logError("GetCarPhoto", err)
		return "", err
	}

	root, err := filepath.Abs(path.Join(a.cfg.Workspace, "media"))
	if err != nil {
		err = fmt.Errorf("resolve media directory: %w", err)
		a.logError("GetCarPhoto", err)
		return "", err
	}

	fullPath := filepath.Join(root, filepath.Clean(photoPath))

	// Ensure the requested file is inside MediaPath.
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		err = fmt.Errorf("resolve photo path: %w", err)
		a.logError("GetCarPhoto", err)
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		err := fmt.Errorf("invalid photo path")
		a.logError("GetCarPhoto", err)
		return "", err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		err = fmt.Errorf("read photo: %w", err)
		a.logError("GetCarPhoto", err)
		return "", err
	}

	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded), nil
}

func (a *App) HasChallongeAPIKey() bool {
	if _, err := a.cfg.Get(config.ChallongeApiKey); err != nil {
		a.logError("HasChallongeAPIKey", err)
		return false
	}
	return true
}

func (a *App) SetChallongeAPIKey(key string) error {
	url := "https://api.challonge.com/v2.1/me.json"

	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("create Challonge API request: %w", err)
		a.logError("SetChallongeAPIKey", err)
		return err
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization-Type", "v1")
	req.Header.Set("Authorization", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("validate Challonge API key: %w", err)
		a.logError("SetChallongeAPIKey", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err = fmt.Errorf("could not validate Challonge API key: HTTP %d", resp.StatusCode)
		a.logError("SetChallongeAPIKey", err)
		return err
	}

	if err := a.cfg.Set(config.ChallongeApiKey, key); err != nil {
		err = fmt.Errorf("save Challonge API key: %w", err)
		a.logError("SetChallongeAPIKey", err)
		return err
	}

	return nil
}

// -- Tournament --

func (a *App) GetTournaments() ([]tournament.Tournament, error) {
	tournaments, err := a.tournamentRepo.GetTournaments()
	if err != nil {
		a.logError("GetTournaments", err)
		return []tournament.Tournament{}, err
	}

	if tournaments == nil {
		return []tournament.Tournament{}, nil
	}

	return tournaments, nil
}

func (a *App) GetTournament(id string) (tournament.Tournament, error) {
	t, err := a.tournamentRepo.GetTournamentById(id)
	if err != nil {
		a.logError("GetTournament", err)
		return tournament.Tournament{}, err
	}

	return t, nil
}

func (a *App) CreateTournament(req tournament.CreateRequest) (tournament.Tournament, error) {
	t, err := a.tournamentRepo.CreateTournament(a.ctx, req)
	if err != nil {
		a.logError("CreateTournament", err)
		return tournament.Tournament{}, err
	}

	log := tournamentactivitylog.CreateTournamentActivityLogInput{
		TournamentID: t.ID,
		Content:      fmt.Sprintf("Tournament %q created", t.Name),
		LogType:      tournamentactivitylog.TOURNAMENT_INIT,
	}

	if err := a.tournamentActivityLogsRepo.Create(a.ctx, log); err != nil {
		a.logError("CreateTournament", err)
		return *t, nil
	}

	return *t, err
}

func (a *App) InsertBulkParticipant(tournamentID string, driverIDs []string) ([]string, error) {
	p, err := a.tournamentRepo.InsertBulkParticipant(a.ctx, tournamentID, driverIDs)
	if err != nil {
		a.logError("InsertBulkParticipant", err)
		return nil, err
	}
	return p, err
}

func (a *App) GetParticipants(tournamentID string) ([]participant.Participant, error) {
	participants, err := a.tournamentRepo.GetParticipants(a.ctx, tournamentID)
	if err != nil {
		a.logError("GetParticipants", err)
		return nil, err
	}
	return participants, err
}

func (a *App) StartTournament(tournamentID string) error {
	err := a.tournamentRepo.StartTournament(a.ctx, tournamentID)
	if err != nil {
		a.logError("StartTournament", err)
		return nil
	}

	log := tournamentactivitylog.CreateTournamentActivityLogInput{
		TournamentID: tournamentID,
		Content:      "The tournament started",
		LogType:      tournamentactivitylog.TOURNAMENT_START,
	}

	return a.tournamentActivityLogsRepo.Create(a.ctx, log)
}

func (a *App) DeleteParticipant(tournamentID, participantID string) error {
	err := a.tournamentRepo.DeleteParticipant(a.ctx, tournamentID, participantID)
	if err != nil {
		a.logError("DeleteParticipant", err)
		return err
	}
	return nil
}

func (a *App) UpdateTournament(id string, req repository.UpdateTournamentRequest) error {
	err := a.tournamentRepo.UpdateTournament(a.ctx, id, req)
	if err != nil {
		a.logError("UpdateTournament", err)
		return err
	}
	return nil
}

func (a *App) UpdateTournamentState(id, state string) error {
	err := a.tournamentRepo.UpdateTournamentState(a.ctx, id, state)
	if err != nil {
		a.logError("UpdateTournamentState", err)
		return err
	}
	return nil
}

func (a *App) DeleteTournament(id, status string) error {
	// only update the tournament state if tournament already been started
	if status == "elimination" {
		if err := a.tournamentRepo.UpdateTournamentState(a.ctx, id, "reset"); err != nil {
			a.logError("DeleteTournament", err)
			return err
		}
	}

	err := a.tournamentRepo.DeleteTournamentById(a.ctx, id)
	if err != nil {
		a.logError("DeleteTournamentByID", err)
		return err
	}
	return nil
}

func (a *App) SyncTournament(id string) error {
	err := a.bracketRepo.Sync(a.ctx, id)
	if err != nil {
		a.logError("SyncTournament", err)
		return err
	}
	return nil
}

func (a *App) ReplaceParticipant(tournamentID, from, to string) error {
	res, err := a.participantsRepo.ReplaceParticipant(a.ctx, from, to)
	if err != nil {
		a.logError("ReplaceParticipant", err)
		return err
	}

	log := tournamentactivitylog.CreateTournamentActivityLogInput{
		TournamentID: tournamentID,
		Content:      fmt.Sprintf("%s took over %s's spot in the bracket", res.To.Name, res.From.Name),
		LogType:      tournamentactivitylog.REPLACE_PARTICIPANTS,
	}

	return a.tournamentActivityLogsRepo.Create(a.ctx, log)
}

// -- Tracks --

func (a *App) GetTracks() ([]track.Track, error) {
	tracks, err := a.tracksRepo.GetAll(a.ctx)
	if err != nil {
		a.logError("GetTracks", err)
		return nil, err
	}
	return tracks, nil
}

func (a *App) CreateTrack(trackreq track.CreateTrackReq) error {
	err := a.tracksRepo.Create(a.ctx, trackreq)
	if err != nil {
		a.logError("DeleteTrack", err)
	}
	return err
}

func (a *App) DeleteTrack(id string) error {
	err := a.tracksRepo.Delete(a.ctx, id)
	if err != nil {
		a.logError("DeleteTrack", err)
	}
	return err
}

func (a *App) UpdateTrack(track track.UpdateTrackReq) error {
	err := a.tracksRepo.Update(a.ctx, track)
	if err != nil {
		a.logError("UpdateTrack", err)
	}
	return err
}

// -- Drivers --

func (a *App) GetDrivers() ([]driver.Driver, error) {
	drivers, err := a.driversRepo.GetAll(a.ctx)
	if err != nil {
		a.logError("GetDrivers", err)
		return nil, err
	}
	return drivers, nil
}

func (a *App) GetDriverByID(id string) (driver.Driver, error) {
	d, err := a.driversRepo.GetByID(a.ctx, id)
	if err != nil {
		a.logError("GetDriverByID", err)
		return driver.Driver{}, err
	}
	return d, nil
}

func (a *App) CreateDriver(req driver.CreateDriverReq) error {
	err := a.driversRepo.Create(a.ctx, req)
	if err != nil {
		a.logError("CreateDriver", err)
	}
	return err
}

func (a *App) UpdateDriver(req driver.UpdateDriverReq) error {
	err := a.driversRepo.Update(a.ctx, req)
	if err != nil {
		a.logError("UpdateDriver", err)
	}
	return err
}

func (a *App) DeleteDriver(id string) error {
	err := a.driversRepo.Delete(a.ctx, id)
	if err != nil {
		a.logError("DeleteDriver", err)
	}
	return err
}

// -- Cars --

func (a *App) UpdateParticipantCar(participantID, registeredCarID string) error {
	err := a.participantsRepo.UpdateRegisteredCar(a.ctx, participantID, registeredCarID)
	if err != nil {
		a.logError("UpdateParticipantCar", err)
	}
	return err
}

func (a *App) CreateRegisteredCar(car car.CreateRegisteredCar) error {
	err := a.registeredCarsRepo.Create(a.ctx, car)
	if err != nil {
		a.logError("CreateRegisteredCar", err)
	}
	return err
}

func (a *App) GetCarByID(id string) (car.RegisteredCar, error) {
	c, err := a.registeredCarsRepo.GetByID(a.ctx, id)
	if err != nil {
		a.logError("GetCarByID", err)
		return car.RegisteredCar{}, err
	}
	return c, nil
}

func (a *App) GetDriverCarList(driverID string) ([]car.RegisteredCar, error) {
	list, err := a.registeredCarsRepo.ListByDriverID(a.ctx, driverID)
	if err != nil {
		a.logError("GetDriverCarList", err)
		return nil, err
	}
	return list, nil
}

func (a *App) ListCarMakes() ([]car.CarMake, error) {
	makes, err := a.registeredCarsRepo.ListCarMakes(a.ctx)
	if err != nil {
		a.logError("ListCarMakes", err)
		return nil, err
	}
	return makes, nil
}

func (a *App) DeleteCar(id string) error {
	err := a.registeredCarsRepo.Delete(a.ctx, id)
	if err != nil {
		a.logError("DeleteCar", err)
	}
	return err
}

func (a *App) UpdateCar(id string, car car.UpdateRegisteredCar) error {
	err := a.registeredCarsRepo.Update(a.ctx, id, car)
	if err != nil {
		a.logError("UpdateCar", err)
	}
	return err
}

// -- Matches --

func (a *App) GetMatch(matchID string) (repository.Match, error) {
	match, err := a.matchesRepo.GetMatch(a.ctx, matchID)
	if err != nil {
		a.logError("GetMatch", err)
		return repository.Match{}, err
	}
	return match, nil
}

func (a *App) GetMatches(tournamentID string) ([][][]repository.Match, error) {
	matches, err := a.matchesRepo.GetMatches(a.ctx, tournamentID)
	if err != nil {
		a.logError("GetMatches", err)
		return nil, err
	}
	return matches, nil
}

func (a *App) UpdateMatchSingleParticipant(req repository.UpdateMatchSingleParticipantParams) error {
	_, err := a.matchesRepo.UpdateMatchSingleParticipant(a.ctx, req)
	if err != nil {
		a.logError("UpdateMatchSingleParticipant", err)
	}
	return err
}

func (a *App) GetActiveMatches(tournamentID string) ([]repository.ActiveBattle, error) {
	matches, err := a.matchesRepo.GetActiveMatches(a.ctx, tournamentID)
	if err != nil {
		a.logError("GetActiveMatches", err)
		return nil, err
	}
	return matches, nil
}

func (a *App) ConcludeBattle(tournamentID, matchID string) error {
	winner, err := a.matchesRepo.ConcludeBattle(a.ctx, tournamentID, matchID)
	if err != nil {
		a.logError("ConcludeBattle", err)
		return err
	}

	log := tournamentactivitylog.CreateTournamentActivityLogInput{
		TournamentID: tournamentID,
		Content:      fmt.Sprintf("%s won with score of %.2f", winner.Name, winner.Score),
		LogType:      tournamentactivitylog.WON_BATTLE,
	}

	err = a.tournamentActivityLogsRepo.Create(a.ctx, log)
	if err != nil {
		a.logError("ConcludeBattle: create activity log", err)
	}
	return err
}

func (a *App) GetNote(matchID string) (string, error) {
	note, err := a.matchesRepo.GetNote(a.ctx, matchID)
	if err != nil {
		a.logError("GetNote", err)
		return "", err
	}
	return note, nil
}

func (a *App) UpdateNote(matchID, text string) error {
	err := a.matchesRepo.UpdateNote(a.ctx, matchID, text)
	if err != nil {
		a.logError("UpdateNote", err)
	}
	return err
}

func (a *App) ResetBattle(tournamentID, matchID string) error {
	err := a.matchesRepo.ResetBattle(a.ctx, tournamentID, matchID)
	if err != nil {
		a.logError("ResetBattle", err)
	}
	return err
}

func (a *App) ChangeLeader(tournamentID, matchID string) error {
	err := a.matchesRepo.ChangeLeader(a.ctx, tournamentID, matchID)
	if err != nil {
		a.logError("ChangeLeader", err)
	}
	return err
}

func (a *App) InitializeMatchScores(param eliminationscore.InitializeScoreParam) ([]eliminationscore.ScoreData, error) {
	scores, err := a.eliminationScoresRepo.InitializeScores(a.ctx, param)
	if err != nil {
		a.logError("InitializeMatchScores", err)
	}
	return scores, err
}

func (a *App) UpdateEliminationScore(param eliminationscore.UpdateEliminationScoreParams) (float64, error) {
	score, err := a.eliminationScoresRepo.UpdateScore(a.ctx, param)
	if err != nil {
		a.logError("UpdateEliminationScore", err)
	}
	return score, err
}

func (a *App) UpdateEliminationScoreByRound(param eliminationscore.UpdateEliminationScoreByRoundParams) (float64, error) {
	score, err := a.eliminationScoresRepo.UpdateScoreByRound(a.ctx, param)
	if err != nil {
		a.logError("UpdateEliminationScore", err)
	}
	return score, err
}

func (a *App) ResetMatchScores(matchID string) error {
	err := a.eliminationScoresRepo.ResetScores(a.ctx, matchID)
	if err != nil {
		a.logError("UpdateEliminationScore", err)
	}
	return err
}

// -- Entrants --

func (a *App) GetEntrants(id string) ([]entrant.Entrant, error) {
	entrants, err := a.entrantsRepo.GetEntrants(a.ctx, id)
	if err != nil {
		a.logError("GetEntrants", err)
		return nil, err
	}
	return entrants, nil
}

func (a *App) UpdateEntrant(req repository.UpdateEntrantParams) error {
	err := a.entrantsRepo.UpdateEntrant(a.ctx, req)
	if err != nil {
		a.logError("UpdateEntrant", err)
	}
	return err
}

func (a *App) ConcludeQualifers(tournamentId string) error {
	err := a.entrantsRepo.ConcludeQualifiers(a.ctx, tournamentId)
	if err != nil {
		a.logError("ConcludeQualifers", err)
		return err
	}

	log := tournamentactivitylog.CreateTournamentActivityLogInput{
		TournamentID: tournamentId,
		Content:      "Qualifying finished",
		LogType:      tournamentactivitylog.QUALIFYING_FINISHED,
	}

	err = a.tournamentActivityLogsRepo.Create(a.ctx, log)
	if err != nil {
		a.logError("ConcludeQualifers: create activity log", err)
	}
	return err
}

// -- Criteria Configs --

func (a *App) GetCriteriaConfigs() ([]repository.CriteriaConfig, error) {
	configs, err := a.criteriaConfigRepo.Get()
	if err != nil {
		a.logError("GetCriteriaConfigs", err)
		return nil, err
	}
	return configs, nil
}

func (a *App) GetCriteriaConfigByID(id int) (repository.CriteriaConfig, error) {
	config, err := a.criteriaConfigRepo.GetByID(id)
	if err != nil {
		a.logError("GetCriteriaConfigByID", err)
		return repository.CriteriaConfig{}, err
	}
	return config, nil
}

func (a *App) CreateCriteriaConfig(config repository.CriteriaConfig) error {
	err := a.criteriaConfigRepo.Create(config)
	if err != nil {
		a.logError("CreateCriteriaConfig", err)
	}
	return err
}

func (a *App) CopyCriteriaConfigById(id int) error {
	_, err := a.criteriaConfigRepo.CopyByID(id)
	if err != nil {
		a.logError("CopyCriteriaConfigById", err)
	}
	return err
}

func (a *App) UpdateCriteriaConfig(id int, config repository.CriteriaConfig) error {
	err := a.criteriaConfigRepo.Update(id, config)
	if err != nil {
		a.logError("UpdateCriteriaConfig", err)
	}
	return err
}

func (a *App) DeleteCriteriaConfig(id int) error {
	err := a.criteriaConfigRepo.Delete(id)
	if err != nil {
		a.logError("DeleteCriteriaConfig", err)
	}
	return err
}

// -- Qualifying Scores --

func (a *App) InitializeQualifyingScores(tournamentID string) error {
	err := a.qualifyingScoresRepo.InitializeScores(a.ctx, tournamentID)
	if err != nil {
		a.logError("InitializeQualifyingScores", err)
	}
	return err
}

func (a *App) GetAllQualifyingScores(tournamentID string) ([]qualifyingscore.Data, error) {
	data, err := a.qualifyingScoresRepo.FindAllScores(a.ctx, tournamentID)
	if err != nil {
		a.logError("GetAllQualifyingScores", err)
		return nil, err
	}
	return data, nil
}

func (a *App) UpdateQualifyingScore(params qualifyingscore.UpdateQualifyingScoreParams) error {
	err := a.qualifyingScoresRepo.UpdateScore(a.ctx, params)
	if err != nil {
		a.logError("UpdateQualifyingScore", err)
	}
	return err
}

// -- Activity Logs --

func (a *App) GetTournamentActivityLogs(tournamentID string) ([]tournamentactivitylog.TournamentActivityLog, error) {
	logs, err := a.tournamentActivityLogsRepo.GetByTournamentID(a.ctx, tournamentID)
	if err != nil {
		a.logError("GetTournamentActivityLogs", err)
		return nil, err
	}
	return logs, nil
}

func (a *App) CreateTournamentActivityLog(input tournamentactivitylog.CreateTournamentActivityLogInput) error {
	err := a.tournamentActivityLogsRepo.Create(a.ctx, input)
	if err != nil {
		a.logError("CreateTournamentActivityLog", err)
	}
	return err
}
