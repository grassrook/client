import type { entrant, tournament } from "$wails/go/models";

// The qualifying order can arrive as an array or a comma separated string
export function parseQualifyingOrder(order: unknown): string[] {
  if (Array.isArray(order)) return order as string[];

  return String(order ?? "")
    .split(",")
    .map((id) => id.trim())
    .filter((id) => id !== "");
}

// Elimination: highest score first.
// Otherwise: the tournament's qualifying order, unknown participants last.
// Ties keep their original relative order.
export function orderEntrants(
  entrants: entrant.Entrant[],
  t: tournament.Tournament,
): entrant.Entrant[] {
  const indexed = entrants.map((e, i) => ({ e, i }));

  if (t.status === "elimination") {
    return indexed
      .sort((a, b) => {
        const scoreA = a.e.score ?? 0;
        const scoreB = b.e.score ?? 0;

        return scoreB - scoreA || a.i - b.i;
      })
      .map(({ e }) => e);
  }

  const rank = new Map(
    parseQualifyingOrder(t.qualifying_order).map((id, i) => [id, i]),
  );

  return indexed
    .sort((a, b) => {
      const ra = rank.get(a.e.participantId) ?? Number.MAX_SAFE_INTEGER;
      const rb = rank.get(b.e.participantId) ?? Number.MAX_SAFE_INTEGER;

      return ra - rb || a.i - b.i;
    })
    .map(({ e }) => e);
}

// Names of the entrants in the order they originally qualified
export function initialOrderNames(
  t: tournament.Tournament,
  entrants: entrant.Entrant[],
): string[] {
  const order = t.initialQualifyingOrder
    ?.split(",")
    .map((id) => id.trim())
    .filter(Boolean);

  if (!order?.length) return [];

  const names = new Map(entrants.map((e) => [e.participantId, e.name]));

  return order
    .map((id) => names.get(id))
    .filter((name): name is string => name !== undefined);
}
