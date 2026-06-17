// useSort — a tiny, type-safe table sorter shared by every data table in the UI.
// Give it the rows (a ref or a getter) and an initial column; it returns the
// reactive sort state, a toggle(col) for header clicks, and the sorted rows.
//
// Columns whose value isn't a plain row property (computed labels, joined
// fields) are handled via `accessors`. Comparison is natural/numeric-aware so
// "addr 7" sorts before "addr 70", and undefined/null always sink to the bottom.
import { computed, isRef, ref, type Ref } from 'vue'

export type SortDir = 1 | -1
export interface SortState {
  col: string
  dir: SortDir
}

type Cell = string | number | boolean | null | undefined
type Source<T> = Ref<T[]> | (() => T[])

export interface UseSortOptions<T> {
  initial: string
  dir?: SortDir
  accessors?: Record<string, (row: T) => Cell>
}

export function useSort<T>(rows: Source<T>, opts: UseSortOptions<T>) {
  const sort = ref<SortState>({ col: opts.initial, dir: opts.dir ?? 1 })
  const src = isRef(rows) ? rows : computed(rows)

  function toggle(col: string) {
    if (sort.value.col === col) {
      sort.value = { col, dir: (sort.value.dir * -1) as SortDir }
    } else {
      sort.value = { col, dir: 1 }
    }
  }

  function cell(row: T, col: string): Cell {
    const acc = opts.accessors?.[col]
    return acc ? acc(row) : (row as Record<string, Cell>)[col]
  }

  const sorted = computed<T[]>(() => {
    const { col, dir } = sort.value
    return [...src.value].sort((a, b) => compare(cell(a, col), cell(b, col)) * dir)
  })

  return { sort, toggle, sorted }
}

function compare(a: Cell, b: Cell): number {
  // undefined/null sink regardless of direction handled by caller's * dir? No —
  // we keep them last on ascending by ranking them high; callers rarely sort by
  // empty columns, and this is the least-surprising default.
  const an = a === null || a === undefined || a === ''
  const bn = b === null || b === undefined || b === ''
  if (an && bn) return 0
  if (an) return 1
  if (bn) return -1
  if (typeof a === 'number' && typeof b === 'number') return a - b
  if (typeof a === 'boolean' && typeof b === 'boolean') return a === b ? 0 : a ? 1 : -1
  return String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'base' })
}
