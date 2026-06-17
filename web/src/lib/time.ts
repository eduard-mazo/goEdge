// Relative-time helpers for source freshness labels ("Actualizado", "últ").
//
// These are hardened against two sentinels the gateway can emit that are valid
// strings but mean "no reading yet":
//   • Go's zero time  → "0001-01-01T00:00:00Z"  (parses to a huge NEGATIVE epoch)
//   • the Unix epoch  → "1970-01-01T00:00:00Z"
// Treating either as a real instant produces nonsense like "hace 17754791h", so
// isRealTime() rejects anything at or before the epoch and the formatters return
// an em dash instead.

// Floor below which a timestamp is considered "never". Go's year-1 zero value is
// negative, so > 0 also screens it out.
const EPOCH_FLOOR_MS = 0

/** True only for a parseable timestamp strictly after the Unix epoch. */
export function isRealTime(iso?: string | null): boolean {
  if (!iso) return false
  const t = new Date(iso).getTime()
  return Number.isFinite(t) && t > EPOCH_FLOOR_MS
}

/**
 * Compact age label: "ahora" / "5s" / "3m" / "2h" / "4d". Returns "—" for a
 * missing/zero timestamp. Pass `now` (e.g. a reactive clock) for live ticking;
 * a future timestamp (clock skew) clamps to "ahora".
 */
export function ago(iso?: string | null, now: number = Date.now()): string {
  if (!isRealTime(iso)) return '—'
  const s = Math.floor((now - new Date(iso as string).getTime()) / 1000)
  if (s < 1) return 'ahora'
  if (s < 60) return s + 's'
  if (s < 3_600) return Math.floor(s / 60) + 'm'
  if (s < 86_400) return Math.floor(s / 3_600) + 'h'
  return Math.floor(s / 86_400) + 'd'
}

/** Spanish "hace …" phrasing for dashboard cards. "—" when never read. */
export function relTime(iso?: string | null, now: number = Date.now()): string {
  const label = ago(iso, now)
  if (label === '—' || label === 'ahora') return label
  return 'hace ' + label
}
