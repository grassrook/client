export namespace car {
	
	export class CarMake {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CarMake(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class CreateRegisteredCar {
	    name: string;
	    description: string;
	    carMakeID: string;
	    driverID: string;
	    photos: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateRegisteredCar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.carMakeID = source["carMakeID"];
	        this.driverID = source["driverID"];
	        this.photos = source["photos"];
	    }
	}
	export class RegisteredCar {
	    id: string;
	    name: string;
	    description: string;
	    carMakeID: string;
	    driverID: string;
	    createdAt: number;
	    updatedAt: number;
	    photos: string[];
	
	    static createFrom(source: any = {}) {
	        return new RegisteredCar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.carMakeID = source["carMakeID"];
	        this.driverID = source["driverID"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.photos = source["photos"];
	    }
	}
	export class UpdateRegisteredCar {
	    name?: string;
	    description?: string;
	    photos?: string[];
	
	    static createFrom(source: any = {}) {
	        return new UpdateRegisteredCar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.photos = source["photos"];
	    }
	}

}

export namespace driver {
	
	export class CreateDriverReq {
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDriverReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class Driver {
	    id: string;
	    name: string;
	    description: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Driver(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class UpdateDriverReq {
	    id: string;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateDriverReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}

}

export namespace entrant {
	
	export class Entrant {
	    id: string;
	    participantId: string;
	    tournamentId: string;
	    name: string;
	    status: string;
	    score?: number;
	    driverId: string;
	    registeredCarID: string;
	
	    static createFrom(source: any = {}) {
	        return new Entrant(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.participantId = source["participantId"];
	        this.tournamentId = source["tournamentId"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.score = source["score"];
	        this.driverId = source["driverId"];
	        this.registeredCarID = source["registeredCarID"];
	    }
	}

}

export namespace participant {
	
	export class Participant {
	    id: string;
	    name: string;
	    tournament_id: string;
	    ChallongeParticipantID?: string;
	    driverID: string;
	    registeredCarID: string;
	    created_at: number;
	    updated_at: number;
	
	    static createFrom(source: any = {}) {
	        return new Participant(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.tournament_id = source["tournament_id"];
	        this.ChallongeParticipantID = source["ChallongeParticipantID"];
	        this.driverID = source["driverID"];
	        this.registeredCarID = source["registeredCarID"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}

}

export namespace qualifyingscore {
	
	export class Result {
	    lead?: number;
	    chase?: number;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lead = source["lead"];
	        this.chase = source["chase"];
	    }
	}
	export class ScoreData {
	    score: number;
	    status: string;
	    results: Record<string, Result>;
	
	    static createFrom(source: any = {}) {
	        return new ScoreData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.status = source["status"];
	        this.results = this.convertValues(source["results"], Result, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Data {
	    participantId: string;
	    scores: ScoreData[];
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.participantId = source["participantId"];
	        this.scores = this.convertValues(source["scores"], ScoreData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UpdateQualifyingScoreParams {
	    TournamentID: string;
	    ParticipantID: string;
	    Round: number;
	    Score: number;
	    Status: string;
	    Results: Record<string, Result>;
	
	    static createFrom(source: any = {}) {
	        return new UpdateQualifyingScoreParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TournamentID = source["TournamentID"];
	        this.ParticipantID = source["ParticipantID"];
	        this.Round = source["Round"];
	        this.Score = source["Score"];
	        this.Status = source["Status"];
	        this.Results = this.convertValues(source["Results"], Result, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace repository {
	
	export class ActiveBattle {
	    id: string;
	    matchId: string;
	    participantId: string;
	    leadStatus: number;
	    rerun: number;
	
	    static createFrom(source: any = {}) {
	        return new ActiveBattle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.matchId = source["matchId"];
	        this.participantId = source["participantId"];
	        this.leadStatus = source["leadStatus"];
	        this.rerun = source["rerun"];
	    }
	}
	export class CriteriaTag {
	    id: string;
	    label: string;
	    type: string;
	    valueType: string;
	    max?: number;
	    multiplier?: number;
	    defaultValue?: number;
	
	    static createFrom(source: any = {}) {
	        return new CriteriaTag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.valueType = source["valueType"];
	        this.max = source["max"];
	        this.multiplier = source["multiplier"];
	        this.defaultValue = source["defaultValue"];
	    }
	}
	export class CriteriaCategory {
	    label: string;
	    tags: CriteriaTag[];
	
	    static createFrom(source: any = {}) {
	        return new CriteriaCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.tags = this.convertValues(source["tags"], CriteriaTag);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CriteriaCategories {
	    common: CriteriaCategory;
	    lead: CriteriaCategory;
	    chase: CriteriaCategory;
	
	    static createFrom(source: any = {}) {
	        return new CriteriaCategories(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.common = this.convertValues(source["common"], CriteriaCategory);
	        this.lead = this.convertValues(source["lead"], CriteriaCategory);
	        this.chase = this.convertValues(source["chase"], CriteriaCategory);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class WeighAllocation {
	    maxScore: number;
	    commonScore: number;
	
	    static createFrom(source: any = {}) {
	        return new WeighAllocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxScore = source["maxScore"];
	        this.commonScore = source["commonScore"];
	    }
	}
	export class ScoreWeighs {
	    lead: WeighAllocation;
	    chase: WeighAllocation;
	
	    static createFrom(source: any = {}) {
	        return new ScoreWeighs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lead = this.convertValues(source["lead"], WeighAllocation);
	        this.chase = this.convertValues(source["chase"], WeighAllocation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CriteriaConfig {
	    id: number;
	    name: string;
	    categories: CriteriaCategories;
	    raw: string;
	    weighMaxScore: boolean;
	    scoreWeighs: ScoreWeighs;
	
	    static createFrom(source: any = {}) {
	        return new CriteriaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.categories = this.convertValues(source["categories"], CriteriaCategories);
	        this.raw = source["raw"];
	        this.weighMaxScore = source["weighMaxScore"];
	        this.scoreWeighs = this.convertValues(source["scoreWeighs"], ScoreWeighs);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Match {
	    matchId: string;
	    participant_name?: string;
	    status: string;
	    round: number;
	    winnerId?: string;
	    challongeParticipantId?: string;
	    participantId?: string;
	    score?: number;
	    bothRunFinished: boolean;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new Match(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.matchId = source["matchId"];
	        this.participant_name = source["participant_name"];
	        this.status = source["status"];
	        this.round = source["round"];
	        this.winnerId = source["winnerId"];
	        this.challongeParticipantId = source["challongeParticipantId"];
	        this.participantId = source["participantId"];
	        this.score = source["score"];
	        this.bothRunFinished = source["bothRunFinished"];
	        this.note = source["note"];
	    }
	}
	
	export class UpdateEntrantParams {
	    TournamentID: string;
	    EntrantID: string;
	    Status?: string;
	    Score?: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateEntrantParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TournamentID = source["TournamentID"];
	        this.EntrantID = source["EntrantID"];
	        this.Status = source["Status"];
	        this.Score = source["Score"];
	    }
	}
	export class UpdateMatchSingleParticipantParams {
	    TournamentID: string;
	    MatchID: string;
	    ParticipantID: string;
	    Score?: number;
	    WinnerID?: string;
	    Status?: string;
	    AppendScore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateMatchSingleParticipantParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TournamentID = source["TournamentID"];
	        this.MatchID = source["MatchID"];
	        this.ParticipantID = source["ParticipantID"];
	        this.Score = source["Score"];
	        this.WinnerID = source["WinnerID"];
	        this.Status = source["Status"];
	        this.AppendScore = source["AppendScore"];
	    }
	}
	export class UpdateTournamentRequest {
	    name?: string;
	    size?: number;
	    startingDate?: string;
	    track?: string;
	    qualifyingOrder?: string[];
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTournamentRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.startingDate = source["startingDate"];
	        this.track = source["track"];
	        this.qualifyingOrder = source["qualifyingOrder"];
	        this.status = source["status"];
	    }
	}

}

export namespace tournament {
	
	export class CreateRequest {
	    name: string;
	    size: number;
	    startingDate: string;
	    trackID: string;
	    qualifyingCriteriaConfigID: number;
	    criteriaConfigID: number;
	    qualifyingRounds: number;
	    randomizeQualifyingOrder: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.startingDate = source["startingDate"];
	        this.trackID = source["trackID"];
	        this.qualifyingCriteriaConfigID = source["qualifyingCriteriaConfigID"];
	        this.criteriaConfigID = source["criteriaConfigID"];
	        this.qualifyingRounds = source["qualifyingRounds"];
	        this.randomizeQualifyingOrder = source["randomizeQualifyingOrder"];
	    }
	}
	export class Tournament {
	    id: string;
	    name: string;
	    starting_date: string;
	    track_id: string;
	    track?: string;
	    qualifying_order: string;
	    status: string;
	    url: string;
	    size: number;
	    progress: number;
	    qualifyingCriteriaConfig?: string;
	    criteriaConfig: string;
	    finalized: boolean;
	    qualifyingRounds: number;
	    randomizeQualifyingOrder: boolean;
	    initialQualifyingOrder?: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Tournament(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.starting_date = source["starting_date"];
	        this.track_id = source["track_id"];
	        this.track = source["track"];
	        this.qualifying_order = source["qualifying_order"];
	        this.status = source["status"];
	        this.url = source["url"];
	        this.size = source["size"];
	        this.progress = source["progress"];
	        this.qualifyingCriteriaConfig = source["qualifyingCriteriaConfig"];
	        this.criteriaConfig = source["criteriaConfig"];
	        this.finalized = source["finalized"];
	        this.qualifyingRounds = source["qualifyingRounds"];
	        this.randomizeQualifyingOrder = source["randomizeQualifyingOrder"];
	        this.initialQualifyingOrder = source["initialQualifyingOrder"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

export namespace tournamentactivitylog {
	
	export class CreateTournamentActivityLogInput {
	    tournament_id: string;
	    content: string;
	    logType: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateTournamentActivityLogInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tournament_id = source["tournament_id"];
	        this.content = source["content"];
	        this.logType = source["logType"];
	    }
	}
	export class TournamentActivityLog {
	    id: string;
	    tournament_id: string;
	    content: string;
	    logType: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new TournamentActivityLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tournament_id = source["tournament_id"];
	        this.content = source["content"];
	        this.logType = source["logType"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

export namespace track {
	
	export class CreateTrackReq {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateTrackReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}
	export class Track {
	    id: string;
	    name: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class UpdateTrackReq {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTrackReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

