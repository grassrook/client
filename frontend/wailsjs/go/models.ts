export namespace entrant {
	
	export class Entrant {
	    id: string;
	    participantId: string;
	    tournamentId: string;
	    name: string;
	    status: string;
	    score?: number;
	
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
	    }
	}

}

export namespace participant {
	
	export class Participant {
	    id: string;
	    name: string;
	    tournament_id: string;
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
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}

}

export namespace qualifyingscore {
	
	export class ScoreData {
	    score: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ScoreData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.status = source["status"];
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
	
	export class CriteriaConfig {
	    id: number;
	    name: string;
	    categories: CriteriaCategories;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new CriteriaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.categories = this.convertValues(source["categories"], CriteriaCategories);
	        this.raw = source["raw"];
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
	    track: string;
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
	        this.track = source["track"];
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
	    criteriaConfig: string;
	    finalized: boolean;
	    qualifyingRounds: number;
	    randomizeQualifyingOrder: boolean;
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
	        this.criteriaConfig = source["criteriaConfig"];
	        this.finalized = source["finalized"];
	        this.qualifyingRounds = source["qualifyingRounds"];
	        this.randomizeQualifyingOrder = source["randomizeQualifyingOrder"];
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

