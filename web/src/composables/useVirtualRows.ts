// useVirtualRows — windowed rendering for long tables (1000+ signals).
//
// Renders only the rows intersecting the viewport plus a small overscan; the
// off-screen rows are represented by two spacer <tr> whose heights reserve the
// real scroll extent. This keeps a single smooth scrollbar and the native
// <table> (so sorting, sticky headers and column alignment all still work)
// while the DOM holds ~30 rows instead of thousands.
//
// REQUIREMENT: every data row must be EXACTLY `rowHeight` tall, or the spacer
// math drifts. Pair with a `.vrow { height: <rowHeight>px }` rule whose cells
// draw their separator with an inset box-shadow (not a border, which would add
// to the row box).
import { computed, onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

export interface VirtualRowsOptions {
  rowHeight: number
  /** Extra rows rendered above/below the viewport to mask fast scrolling. */
  overscan?: number
}

export interface VirtualRow<T> {
  row: T
  index: number
}

export function useVirtualRows<T>(
  source: Ref<T[]>,
  viewport: Ref<HTMLElement | null>,
  opts: VirtualRowsOptions,
) {
  const overscan = opts.overscan ?? 8
  const scrollTop = ref(0)
  const viewportH = ref(0)

  const onScroll = () => {
    const el = viewport.value
    if (el) scrollTop.value = el.scrollTop
  }
  const measure = () => {
    const el = viewport.value
    if (el) viewportH.value = el.clientHeight
  }

  let ro: ResizeObserver | undefined
  onMounted(() => {
    const el = viewport.value
    if (!el) return
    el.addEventListener('scroll', onScroll, { passive: true })
    measure()
    if (typeof ResizeObserver !== 'undefined') {
      ro = new ResizeObserver(measure)
      ro.observe(el)
    }
  })
  onBeforeUnmount(() => {
    viewport.value?.removeEventListener('scroll', onScroll)
    ro?.disconnect()
  })

  // When the set shrinks (e.g. a filter narrows 1000 → 12 rows), the viewport can
  // be stranded past the new end — snap it back to a valid offset.
  watch(
    () => source.value.length,
    (n) => {
      const el = viewport.value
      if (!el) return
      const max = Math.max(0, n * opts.rowHeight - el.clientHeight)
      if (el.scrollTop > max) {
        el.scrollTop = max
        scrollTop.value = max
      }
    },
  )

  const total = computed(() => source.value.length)
  const start = computed(() =>
    Math.max(0, Math.floor(scrollTop.value / opts.rowHeight) - overscan),
  )
  const perView = computed(() => Math.ceil((viewportH.value || 1) / opts.rowHeight) + overscan * 2)
  const end = computed(() => Math.min(total.value, start.value + perView.value))

  const visible = computed<VirtualRow<T>[]>(() =>
    source.value.slice(start.value, end.value).map((row, i) => ({ row, index: start.value + i })),
  )
  const padTop = computed(() => start.value * opts.rowHeight)
  const padBottom = computed(() => Math.max(0, (total.value - end.value) * opts.rowHeight))

  return { visible, padTop, padBottom, start, end, total }
}
