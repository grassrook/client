import type { repository } from "$wails/go/models";

type Result = {
  lead?: number;
  chase?: number;
};

// sum the max of every non-deduction tag in a category.
function sumMax(tags?: repository.CriteriaTag[]): number {
  return (tags ?? [])
    .filter((tag) => tag.type !== "deduction")
    .reduce((sum, tag) => sum + tag.max!, 0);
}

// converts a score from one scale to another while keeping the same proportion.
// 1. Find the proportion earned: raw / fromMax
// 2. Apply that proportion to the new maximum: proportion * toMax
//
// Example:
//   - raw judged score is 30
//   - original max (before fix weight) is 45
//   - new weighted max score for this criteria is 50
//
//   proportion = 30 / 45 = 0.667 (66.7% of the original max was earned)
//   result     = 0.667 * 50 = 33.33
//
// For judged scores, pass (maxScore - commonScore) as toMax.
// For common scores, pass commonScore as toMax.
// Returns 0 if either max is zero or negative (divide-by-zero guard, and
// what you get when the config hasn't loaded yet)
function scale(raw: number, fromMax: number, toMax: number): number {
  if (fromMax <= 0 || toMax <= 0) return 0;
  return (raw / fromMax) * toMax;
}

function roundToNearestTenths(n: number): number {
  return Math.round(n * 10) / 10;
}

class Judgement {
  processing = $state(false);
  results: Record<string, Result> = $state({});

  weighMaxScore = $state(false);
  scoreWeighs: repository.ScoreWeighs | undefined = $state();
  originalMax = $state({ lead: 0, chase: 0, common: 0 });

  set(id: string, state: keyof Result, value: number) {
    this.results = {
      ...this.results,
      [id]: {
        ...this.results[id],
        [state]: value,
      },
    };
  }

  get(id: string) {
    return this.results[id];
  }

  setUseWeighMaxScore(value: boolean) {
    this.weighMaxScore = value;
  }

  // sets the cap flag, the caps, and the original max in one go from the criteria config.
  setConfig(config: repository.CriteriaConfig) {
    this.weighMaxScore = config.weighMaxScore ?? false;
    this.scoreWeighs = config.scoreWeighs;

    // find the summed score of these 3 categories as their 100% proportion
    this.originalMax = {
      lead: sumMax(config.categories?.lead?.tags),
      chase: sumMax(config.categories?.chase?.tags),
      common: sumMax(config.categories?.common?.tags),
    };
  }

  finalizedScore() {
    const raw = {
      lead: { common: 0, judged: 0 },
      chase: { common: 0, judged: 0 },
    };

    // aggegates the score given by the judge by separating common values and
    // and the main score (lead or chase)
    for (const [id, judgement] of Object.entries(this.results)) {
      const part = id.startsWith("common_") ? "common" : "judged";
      raw.lead[part] += judgement.lead ?? 0;
      raw.chase[part] += judgement.chase ?? 0;
    }

    // skip readjusting if we dont use score caps
    if (!this.weighMaxScore || !this.scoreWeighs) {
      return {
        lead: raw.lead.common + raw.lead.judged,
        chase: raw.chase.common + raw.chase.judged,
      };
    }

    const { lead: l, chase: c } = this.scoreWeighs;
    const m = this.originalMax;

    // returns readjusted scores based on their original proportion to new max score cap
    return {
      lead: roundToNearestTenths(
        scale(raw.lead.common, m.common, l.commonScore) +
          scale(raw.lead.judged, m.lead, l.maxScore - l.commonScore),
      ), // lead proportion is derived from maxScoreCap - commonScore
      chase: roundToNearestTenths(
        scale(raw.chase.common, m.common, c.commonScore) +
          scale(raw.chase.judged, m.chase, c.maxScore - c.commonScore),
      ),
    };
  }

  clear() {
    this.results = {};
  }
}

export { Judgement };
