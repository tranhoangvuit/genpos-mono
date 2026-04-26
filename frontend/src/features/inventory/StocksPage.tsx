import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Filter,
  Package,
  Search,
  Warehouse,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { AdjustOnHandPopover, type AdjustReason } from './AdjustOnHandPopover'
import { useStockOnHand, useStores, useVariantPicker } from './hooks'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const RED = 'hsl(0 74% 42%)'
const LOW_INK = 'hsl(35 85% 35%)'

const PAGE_SIZE = 25

type SortKey = 'product' | 'sku' | 'onhand' | 'committed'
type SortDir = 'asc' | 'desc'

type Row = {
  variantId: string
  productName: string
  variantName: string
  sku: string
  onHand: number
}

type AdjustTarget = {
  variantId: string
  productLabel: string
  from: number
  to: number
  anchor: DOMRect
}

export function StocksPage() {
  const { t } = useTranslation()
  const { data: stores } = useStores()
  const { data: variants } = useVariantPicker()

  const [storeId, setStoreId] = useState<string>('')
  const [storeMenuOpen, setStoreMenuOpen] = useState(false)
  const storeMenuRef = useRef<HTMLDivElement>(null)

  const { data: onHandRows, isLoading } = useStockOnHand(storeId)

  // Local optimistic overrides applied after the user saves an adjustment.
  // Persists only for the page lifetime — backend AdjustStock RPC is TODO.
  const [overrides, setOverrides] = useState<Record<string, number>>({})

  const onHandByVariant = useMemo(() => {
    const m = new Map<string, number>()
    for (const r of onHandRows ?? []) {
      const n = Number(r.on_hand)
      if (Number.isFinite(n)) m.set(r.variant_id, n)
    }
    return m
  }, [onHandRows])

  const rows = useMemo<Row[]>(() => {
    return (variants ?? []).map((v) => ({
      variantId: v.id,
      productName: v.productName,
      variantName: v.variantName,
      sku: v.sku,
      onHand: overrides[v.id] ?? onHandByVariant.get(v.id) ?? 0,
    }))
  }, [variants, onHandByVariant, overrides])

  const [query, setQuery] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [sortKey, setSortKey] = useState<SortKey>('product')
  const [sortDir, setSortDir] = useState<SortDir>('asc')
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [adjust, setAdjust] = useState<AdjustTarget | null>(null)
  const sortRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
      if (!storeMenuRef.current?.contains(e.target as Node)) setStoreMenuOpen(false)
    }
    if (sortOpen || storeMenuOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen, storeMenuOpen])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let arr = rows
    if (q) {
      arr = arr.filter(
        (r) =>
          r.productName.toLowerCase().includes(q) ||
          r.variantName.toLowerCase().includes(q) ||
          r.sku.toLowerCase().includes(q),
      )
    }
    arr = [...arr].sort((a, b) => {
      let cmp = 0
      switch (sortKey) {
        case 'sku':
          cmp = a.sku.localeCompare(b.sku)
          break
        case 'onhand':
          cmp = a.onHand - b.onHand
          break
        case 'committed':
          cmp = 0
          break
        default:
          cmp = a.productName.localeCompare(b.productName)
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return arr
  }, [rows, query, sortKey, sortDir])

  const pageStart = page * PAGE_SIZE
  const pageRows = filtered.slice(pageStart, pageStart + PAGE_SIZE)
  const pageEnd = pageStart + pageRows.length
  const allSelectedOnPage =
    pageRows.length > 0 && pageRows.every((r) => selected.has(r.variantId))

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }
  function toggleAll() {
    setSelected((prev) => {
      const next = new Set(prev)
      if (allSelectedOnPage) pageRows.forEach((r) => next.delete(r.variantId))
      else pageRows.forEach((r) => next.add(r.variantId))
      return next
    })
  }

  function openAdjust(row: Row, to: number, input: HTMLInputElement) {
    if (to === row.onHand) return
    const label =
      row.variantName && row.variantName !== row.productName
        ? `${row.productName} · ${row.variantName}`
        : row.productName
    setAdjust({
      variantId: row.variantId,
      productLabel: label,
      from: row.onHand,
      to,
      anchor: input.getBoundingClientRect(),
    })
  }

  function commitAdjust(_reason: AdjustReason, _note: string) {
    if (!adjust) return
    // TODO: wire backend AdjustStock RPC. For now, store locally so the saved
    // value sticks until the page is reloaded.
    setOverrides((prev) => ({ ...prev, [adjust.variantId]: adjust.to }))
    setAdjust(null)
  }

  function cancelAdjust() {
    setAdjust(null)
  }

  const selectedStoreName = storeId
    ? stores?.find((s) => s.id === storeId)?.name
    : t('inventory.allStockIns', 'All stores')

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <Warehouse className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('inventory.stocks', 'Stocks')}
          <span
            className="rounded-md px-2 py-[2px] text-[12.5px] font-medium"
            style={{ background: MUTED, color: MUTED_FG }}
          >
            {filtered.length}
          </span>
        </h1>
        <div className="flex items-center gap-2">
          <div className="relative" ref={storeMenuRef}>
            <button
              type="button"
              onClick={() => setStoreMenuOpen((v) => !v)}
              className="inline-flex h-9 items-center gap-1.5 rounded-md border bg-white px-3 text-[13px] font-medium hover:bg-[hsl(210_40%_96%)]"
              style={{ borderColor: BORDER, color: FG }}
            >
              {selectedStoreName ?? t('inventory.store', 'Store')}
              <ChevronDown className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
            </button>
            {storeMenuOpen && (
              <div
                className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[200px] rounded-md border bg-white p-1 shadow-lg"
                style={{ borderColor: BORDER }}
              >
                <StoreMenuItem
                  label={t('inventory.allStockIns', 'All stores')}
                  active={storeId === ''}
                  onClick={() => {
                    setStoreId('')
                    setStoreMenuOpen(false)
                  }}
                />
                {(stores ?? []).map((s) => (
                  <StoreMenuItem
                    key={s.id}
                    label={s.name}
                    active={storeId === s.id}
                    onClick={() => {
                      setStoreId(s.id)
                      setStoreMenuOpen(false)
                    }}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      <section
        className="overflow-hidden rounded-lg border bg-white"
        style={{ borderColor: BORDER }}
      >
        <div
          className="flex items-center justify-between border-b px-2.5 py-1.5"
          style={{ borderColor: BORDER }}
        >
          <div
            className="flex items-center gap-0.5 px-1 text-[13px] font-semibold"
            style={{ color: FG }}
          >
            {t('inventory.allStocks', 'All stocks')}
            <span
              className="ml-2 rounded-md px-2 py-[2px] text-[12px] font-medium"
              style={{ background: MUTED, color: MUTED_FG }}
            >
              {filtered.length}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <ToolBtn
              active={filterOpen}
              onClick={() => setFilterOpen((v) => !v)}
              aria-label="Search"
            >
              <Search className="h-[15px] w-[15px]" strokeWidth={2} />
            </ToolBtn>
            <ToolBtn
              active={filterOpen}
              onClick={() => setFilterOpen((v) => !v)}
              aria-label="Filter"
            >
              <Filter className="h-[15px] w-[15px]" strokeWidth={2} />
            </ToolBtn>
            <div className="relative" ref={sortRef}>
              <ToolBtn
                active={sortOpen}
                onClick={() => setSortOpen((v) => !v)}
                aria-label="Sort"
              >
                <ArrowDownUp className="h-[15px] w-[15px]" strokeWidth={2} />
              </ToolBtn>
              {sortOpen && (
                <SortPopover
                  sortKey={sortKey}
                  sortDir={sortDir}
                  onKey={setSortKey}
                  onDir={setSortDir}
                />
              )}
            </div>
          </div>
        </div>

        {filterOpen && (
          <div
            className="border-b bg-white px-3 pb-3 pt-2.5"
            style={{ borderColor: BORDER }}
          >
            <div className="flex items-center gap-2.5">
              <div
                className="flex h-9 flex-1 items-center gap-2 rounded-md border bg-white px-3"
                style={{
                  borderColor: BLUE,
                  boxShadow: `0 0 0 3px hsl(221 83% 53% / 0.15)`,
                }}
              >
                <Search
                  className="h-[15px] w-[15px]"
                  strokeWidth={2}
                  style={{ color: MUTED_FG }}
                />
                <input
                  autoFocus
                  placeholder={t('inventory.searchStocks', 'Search by product or SKU')}
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value)
                    setPage(0)
                  }}
                  className="flex-1 border-none bg-transparent text-[13px] outline-none"
                />
              </div>
              <button
                type="button"
                onClick={() => {
                  setFilterOpen(false)
                  setQuery('')
                  setPage(0)
                }}
                className="rounded-md px-2.5 py-2 text-[13px] font-medium hover:bg-[hsl(210_40%_96%)]"
              >
                {t('common.cancel', 'Cancel')}
              </button>
            </div>
          </div>
        )}

        <div className="w-full overflow-x-auto">
          <table className="w-full border-collapse" style={{ minWidth: 1040 }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 48 }} />
              <col />
              <col style={{ width: 160 }} />
              <col style={{ width: 110 }} />
              <col style={{ width: 110 }} />
              <col style={{ width: 150 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th />
                <Th>{t('inventory.product', 'Product')}</Th>
                <Th>{t('inventory.sku', 'SKU')}</Th>
                <Th align="right">{t('inventory.unavailable', 'Unavailable')}</Th>
                <Th align="right">{t('inventory.committed', 'Committed')}</Th>
                <Th align="right">{t('inventory.onHand', 'On hand')}</Th>
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td
                    colSpan={7}
                    className="py-10 text-center text-[13px]"
                    style={{ color: MUTED_FG }}
                  >
                    {t('common.loading', 'Loading…')}
                  </td>
                </tr>
              )}
              {!isLoading && pageRows.length === 0 && (
                <tr>
                  <td colSpan={7} className="py-12">
                    <div className="flex flex-col items-center gap-2 text-center">
                      <div
                        className="grid h-10 w-10 place-items-center rounded-md"
                        style={{ background: MUTED, color: MUTED_FG }}
                      >
                        <Package className="h-5 w-5" />
                      </div>
                      <div className="text-[13.5px] font-semibold" style={{ color: FG }}>
                        {t('inventory.noStocks', 'No stocks yet.')}
                      </div>
                      <div className="text-[12.5px]" style={{ color: MUTED_FG }}>
                        {t(
                          'inventory.noStocksHint',
                          'Stock levels appear once products are received or counted.',
                        )}
                      </div>
                    </div>
                  </td>
                </tr>
              )}
              {pageRows.map((r) => {
                const productLabel =
                  r.variantName && r.variantName !== r.productName
                    ? `${r.productName} · ${r.variantName}`
                    : r.productName
                const lowStock = r.onHand > 0 && r.onHand <= 5
                const oversold = r.onHand < 0
                return (
                  <tr
                    key={r.variantId}
                    className="transition hover:bg-[hsl(210_40%_96%_/_0.3)]"
                  >
                    <Td>
                      <Check
                        checked={selected.has(r.variantId)}
                        onClick={() => toggle(r.variantId)}
                      />
                    </Td>
                    <Td>
                      <div
                        className="grid h-7 w-7 flex-shrink-0 place-items-center rounded-md"
                        style={{
                          background: MUTED,
                          border: `1px solid ${BORDER}`,
                          color: MUTED_FG,
                        }}
                      >
                        <Package className="h-3.5 w-3.5" />
                      </div>
                    </Td>
                    <Td>
                      <div className="flex min-w-0 flex-col gap-[3px]">
                        <span className="truncate font-medium" style={{ color: FG }}>
                          {productLabel}
                        </span>
                        {(lowStock || oversold) && (
                          <span
                            className="inline-flex w-fit items-center rounded px-1.5 py-[1px] text-[11.5px] font-medium"
                            style={{
                              background: MUTED,
                              color: oversold ? RED : LOW_INK,
                            }}
                          >
                            {oversold
                              ? t('inventory.outOfStock', 'Out of stock')
                              : t('inventory.lowStock', 'Low stock')}
                          </span>
                        )}
                      </div>
                    </Td>
                    <Td>
                      <span
                        className="text-[12.5px]"
                        style={{
                          color: FG,
                          fontFamily: '"JetBrains Mono", ui-monospace, monospace',
                        }}
                      >
                        {r.sku || '—'}
                      </span>
                    </Td>
                    <Td align="right">
                      <ZeroFaint n={0} />
                    </Td>
                    <Td align="right">
                      <ZeroFaint n={0} />
                    </Td>
                    <Td align="right">
                      <OnHandStepper
                        value={r.onHand}
                        onCommit={(next, input) => openAdjust(r, next, input)}
                      />
                    </Td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <div
          className="flex items-center gap-2.5 border-t px-4 py-3 text-[12.5px]"
          style={{
            borderColor: BORDER,
            background: 'hsl(210 40% 96.1% / 0.3)',
            color: MUTED_FG,
          }}
        >
          <PageNavBtn
            disabled={page === 0}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
          >
            <ChevronLeft className="h-3.5 w-3.5" strokeWidth={2} />
          </PageNavBtn>
          <span>
            {filtered.length === 0 ? '0' : `${pageStart + 1} – ${pageEnd}`} of{' '}
            {filtered.length}
          </span>
          <div className="flex-1" />
          <PageNavBtn
            disabled={pageEnd >= filtered.length}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
            <ChevronRight className="h-3.5 w-3.5" strokeWidth={2} />
          </PageNavBtn>
        </div>
      </section>

      {adjust && (
        <AdjustOnHandPopover
          productLabel={adjust.productLabel}
          from={adjust.from}
          to={adjust.to}
          anchor={adjust.anchor}
          onCancel={cancelAdjust}
          onSave={commitAdjust}
        />
      )}
    </div>
  )
}

function OnHandStepper({
  value,
  onCommit,
}: {
  value: number
  onCommit: (next: number, input: HTMLInputElement) => void
}) {
  const [draft, setDraft] = useState<string>(String(value))
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    setDraft(String(value))
  }, [value])

  function commit() {
    const n = Number(draft)
    if (!Number.isFinite(n) || n === value) {
      setDraft(String(value))
      return
    }
    if (inputRef.current) onCommit(n, inputRef.current)
  }

  function step(delta: number) {
    const n = Number(draft || 0) + delta
    setDraft(String(n))
    if (inputRef.current) onCommit(n, inputRef.current)
  }

  const neg = Number(draft) < 0

  return (
    <div
      className="float-right inline-flex h-7 items-center rounded-md border bg-white pl-2.5 pr-1 transition focus-within:shadow-[0_0_0_2px_hsl(221_83%_53%_/_0.15)]"
      style={{
        borderColor: BORDER,
        minWidth: 110,
        maxWidth: 140,
        fontVariantNumeric: 'tabular-nums',
      }}
    >
      <input
        ref={inputRef}
        type="text"
        inputMode="numeric"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.currentTarget.blur()
          }
        }}
        className="w-full border-none bg-transparent p-0 text-[13px] outline-none"
        style={{
          color: neg ? RED : FG,
          fontVariantNumeric: 'tabular-nums',
        }}
      />
      <div className="ml-1 flex flex-col">
        <button
          type="button"
          tabIndex={-1}
          onClick={() => step(1)}
          aria-label="Increase"
          className="grid h-3 w-[18px] cursor-pointer place-items-center rounded-[3px] hover:bg-[hsl(210_40%_96%)]"
          style={{ color: MUTED_FG }}
        >
          <ArrowUp className="h-2.5 w-2.5" strokeWidth={2.4} />
        </button>
        <button
          type="button"
          tabIndex={-1}
          onClick={() => step(-1)}
          aria-label="Decrease"
          className="grid h-3 w-[18px] cursor-pointer place-items-center rounded-[3px] hover:bg-[hsl(210_40%_96%)]"
          style={{ color: MUTED_FG }}
        >
          <ArrowDown className="h-2.5 w-2.5" strokeWidth={2.4} />
        </button>
      </div>
    </div>
  )
}

function StoreMenuItem({
  label,
  active,
  onClick,
}: {
  label: string
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center justify-between rounded-[5px] px-2.5 py-1.5 text-left text-[13px] hover:bg-[hsl(210_40%_96%)]"
      style={{ color: active ? BLUE_INK : FG, fontWeight: active ? 600 : 400 }}
    >
      {label}
    </button>
  )
}

function ToolBtn({
  children,
  active,
  onClick,
  ...rest
}: {
  children: React.ReactNode
  active?: boolean
  onClick?: () => void
} & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border bg-white transition hover:bg-[hsl(210_40%_96%)]"
      style={{
        borderColor: active ? `hsl(221 83% 53% / 0.3)` : BORDER,
        background: active ? `hsl(221 83% 53% / 0.1)` : 'white',
        color: active ? BLUE_INK : MUTED_FG,
      }}
      {...rest}
    >
      {children}
    </button>
  )
}

function SortPopover({
  sortKey,
  sortDir,
  onKey,
  onDir,
}: {
  sortKey: SortKey
  sortDir: SortDir
  onKey: (k: SortKey) => void
  onDir: (d: SortDir) => void
}) {
  const items: Array<[SortKey, string]> = [
    ['product', 'Product'],
    ['sku', 'SKU'],
    ['onhand', 'On hand'],
    ['committed', 'Committed'],
  ]
  return (
    <div
      className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[240px] rounded-lg border bg-white p-2"
      style={{
        borderColor: BORDER,
        boxShadow:
          '0 16px 40px -12px hsl(222 47% 11% / 0.18), 0 4px 12px hsl(222 47% 11% / 0.06)',
      }}
    >
      <div
        className="px-2 pb-1 pt-1.5 text-[11.5px] font-semibold uppercase tracking-[0.06em]"
        style={{ color: MUTED_FG }}
      >
        Sort by
      </div>
      {items.map(([k, label]) => {
        const on = sortKey === k
        return (
          <button
            key={k}
            type="button"
            onClick={() => onKey(k)}
            className="flex w-full cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-[13px] transition hover:bg-[hsl(210_40%_96%)]"
            style={{ color: FG }}
          >
            <span
              className="grid h-4 w-4 flex-shrink-0 place-items-center rounded-full border-[1.5px]"
              style={{ borderColor: on ? BLUE : BORDER }}
            >
              {on && (
                <span className="h-2 w-2 rounded-full" style={{ background: BLUE }} />
              )}
            </span>
            {label}
          </button>
        )
      })}
      <div className="my-1.5 h-px" style={{ background: BORDER }} />
      {(['asc', 'desc'] as SortDir[]).map((d) => {
        const on = sortDir === d
        const Icon = d === 'asc' ? ArrowUp : ArrowDown
        return (
          <button
            key={d}
            type="button"
            onClick={() => onDir(d)}
            className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-[13px] transition hover:bg-[hsl(210_40%_96%)]"
            style={{
              color: on ? BLUE_INK : FG,
              fontWeight: on ? 600 : 400,
            }}
          >
            <Icon className="h-[13px] w-[13px]" strokeWidth={2.2} />
            {d === 'asc' ? 'Ascending' : 'Descending'}
          </button>
        )
      })}
    </div>
  )
}

function Th({
  children,
  align = 'left',
}: {
  children?: React.ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <th
      className="whitespace-nowrap px-3 py-2.5 text-[12px] font-medium"
      style={{
        textAlign: align,
        color: MUTED_FG,
        background: 'hsl(210 40% 96.1% / 0.4)',
        borderBottom: `1px solid ${BORDER}`,
      }}
    >
      {children}
    </th>
  )
}

function Td({
  children,
  align = 'left',
}: {
  children?: React.ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <td
      className="px-3 py-3 text-[13px] align-middle"
      style={{
        textAlign: align,
        color: FG,
        borderBottom: `1px solid ${BORDER}`,
      }}
    >
      {children}
    </td>
  )
}

function Check({
  checked,
  onClick,
}: {
  checked: boolean
  onClick: () => void
}) {
  return (
    <span
      role="checkbox"
      aria-checked={checked}
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      className="inline-grid h-4 w-4 cursor-pointer place-items-center rounded border-[1.5px] transition-colors"
      style={{
        borderColor: checked ? BLUE : BORDER,
        background: checked ? BLUE : 'white',
        color: 'white',
      }}
    >
      {checked && (
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth={3.5}
          strokeLinecap="round"
          strokeLinejoin="round"
          className="h-[11px] w-[11px]"
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
      )}
    </span>
  )
}

function ZeroFaint({ n }: { n: number }) {
  if (n === 0) {
    return (
      <span className="tabular-nums" style={{ color: MUTED_FG }}>
        0
      </span>
    )
  }
  return (
    <span className="tabular-nums" style={{ color: FG }}>
      {n}
    </span>
  )
}

function PageNavBtn({
  children,
  disabled,
  onClick,
}: {
  children: React.ReactNode
  disabled?: boolean
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-white px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)] disabled:cursor-not-allowed disabled:opacity-50"
      style={{ borderColor: BORDER, color: FG }}
    >
      {children}
    </button>
  )
}
