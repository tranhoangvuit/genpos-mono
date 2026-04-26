import { useNavigate } from '@tanstack/react-router'
import { useQuery as usePowerSyncQuery } from '@powersync/react'
import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  ChevronLeft,
  ChevronRight,
  Filter,
  Package,
  Plus,
  Search,
  Warehouse,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'

import { useStores, useVariantPicker } from './hooks'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'

const PAGE_SIZE = 15

type StockMovementRow = {
  id: string
  store_id: string
  variant_id: string
  direction: string
  quantity: string
  movement_type: string
  reference_type: string | null
  reference_id: string | null
  created_at: string
}

type SortKey = 'date' | 'store' | 'reason' | 'qty'
type SortDir = 'asc' | 'desc'

const REASON_LABEL: Record<string, string> = {
  purchase: 'Purchase order',
  stock_in: 'Receive without PO',
  transfer_in: 'Transfer in',
  stock_take: 'Stock take',
  adjustment: 'Adjustment',
  refund: 'Refund',
  sale: 'Sale',
  transfer_out: 'Transfer out',
}

function formatDate(iso: string) {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return '—'
    return d.toLocaleDateString('en-US', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    })
  } catch {
    return '—'
  }
}

export function StockInsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const { data: stores } = useStores()
  const { data: variants } = useVariantPicker()

  const { data: rows, isLoading } = usePowerSyncQuery<StockMovementRow>(
    `SELECT id, store_id, variant_id, direction, quantity, movement_type,
            reference_type, reference_id, created_at
     FROM stock_movements
     WHERE direction = 'in'
     ORDER BY created_at DESC`,
  )

  const storeName = useMemo(
    () => new Map((stores ?? []).map((s) => [s.id, s.name])),
    [stores],
  )
  const variantById = useMemo(
    () => new Map((variants ?? []).map((v) => [v.id, v])),
    [variants],
  )

  const [query, setQuery] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [sortKey, setSortKey] = useState<SortKey>('date')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const sortRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
    }
    if (sortOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen])

  const all = rows ?? []
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let arr = all
    if (q) {
      arr = arr.filter((r) => {
        const v = variantById.get(r.variant_id)
        const sName = storeName.get(r.store_id) ?? ''
        return (
          (v?.productName.toLowerCase().includes(q) ?? false) ||
          (v?.variantName.toLowerCase().includes(q) ?? false) ||
          (v?.sku.toLowerCase().includes(q) ?? false) ||
          sName.toLowerCase().includes(q) ||
          (r.reference_id?.toLowerCase().includes(q) ?? false)
        )
      })
    }
    arr = [...arr].sort((a, b) => {
      let cmp = 0
      switch (sortKey) {
        case 'store':
          cmp = (storeName.get(a.store_id) ?? '').localeCompare(
            storeName.get(b.store_id) ?? '',
          )
          break
        case 'reason':
          cmp = a.movement_type.localeCompare(b.movement_type)
          break
        case 'qty':
          cmp = Number(a.quantity) - Number(b.quantity)
          break
        default:
          cmp = a.created_at.localeCompare(b.created_at)
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return arr
  }, [all, query, sortKey, sortDir, storeName, variantById])

  const pageStart = page * PAGE_SIZE
  const pageRows = filtered.slice(pageStart, pageStart + PAGE_SIZE)
  const pageEnd = pageStart + pageRows.length
  const allSelectedOnPage =
    pageRows.length > 0 && pageRows.every((r) => selected.has(r.id))

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
      if (allSelectedOnPage) pageRows.forEach((r) => next.delete(r.id))
      else pageRows.forEach((r) => next.add(r.id))
      return next
    })
  }

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <Warehouse className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('nav.stockIns', 'Stock ins')}
        </h1>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() =>
              void navigate({
                to: '/$subdomain/inventory/stock-ins/new',
                params: { subdomain },
              })
            }
            className="inline-flex h-9 items-center gap-1.5 rounded-md px-3.5 text-[13px] font-medium text-white transition"
            style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
            }}
          >
            <Plus className="h-3.5 w-3.5" strokeWidth={2.2} />
            {t('inventory.newStockIn', 'New stock in')}
          </button>
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
          <div className="flex items-center gap-0.5 px-1 text-[13px] font-semibold" style={{ color: FG }}>
            {t('inventory.allStockIns', 'All stock ins')}
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
                  placeholder={t('inventory.searchStockIns', 'Search stock ins')}
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
          <table className="w-full border-collapse" style={{ minWidth: 980 }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 160 }} />
              <col />
              <col style={{ width: 160 }} />
              <col style={{ width: 160 }} />
              <col style={{ width: 110 }} />
              <col style={{ width: 130 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th>{t('inventory.date', 'Date')}</Th>
                <Th>{t('inventory.products', 'Product')}</Th>
                <Th>{t('inventory.store', 'Store')}</Th>
                <Th>{t('inventory.reason', 'Reason')}</Th>
                <Th align="right">{t('inventory.qty', 'Quantity')}</Th>
                <Th>{t('inventory.reference', 'Reference')}</Th>
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
                        {t('inventory.noStockIns', 'No stock ins yet')}
                      </div>
                      <div className="text-[12.5px]" style={{ color: MUTED_FG }}>
                        {t(
                          'inventory.noStockInsHint',
                          'Receiving stock or creating a manual stock in will appear here.',
                        )}
                      </div>
                    </div>
                  </td>
                </tr>
              )}
              {pageRows.map((r) => {
                const v = variantById.get(r.variant_id)
                const productLabel = v
                  ? v.variantName && v.variantName !== v.productName
                    ? `${v.productName} · ${v.variantName}`
                    : v.productName
                  : r.variant_id
                const reason = REASON_LABEL[r.movement_type] ?? r.movement_type
                return (
                  <tr key={r.id} className="transition hover:bg-[hsl(210_40%_96%_/_0.3)]">
                    <Td>
                      <Check
                        checked={selected.has(r.id)}
                        onClick={() => toggle(r.id)}
                      />
                    </Td>
                    <Td>
                      <span className="font-medium" style={{ color: FG }}>
                        {formatDate(r.created_at)}
                      </span>
                    </Td>
                    <Td>
                      <div className="flex items-center gap-2">
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
                        <div className="flex min-w-0 flex-col">
                          <span className="truncate font-medium" style={{ color: FG }}>
                            {productLabel}
                          </span>
                          {v?.sku && (
                            <span
                              className="truncate text-[11.5px]"
                              style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                            >
                              {v.sku}
                            </span>
                          )}
                        </div>
                      </div>
                    </Td>
                    <Td>{storeName.get(r.store_id) ?? '—'}</Td>
                    <Td>
                      <span
                        className="inline-flex items-center rounded-full px-2.5 py-[2px] text-[12px] font-medium"
                        style={{ background: DONE_SOFT, color: DONE_INK }}
                      >
                        {reason}
                      </span>
                    </Td>
                    <Td align="right">
                      <span className="font-medium tabular-nums" style={{ color: FG }}>
                        +{Number(r.quantity).toLocaleString()}
                      </span>
                    </Td>
                    <Td>
                      <span
                        className="text-[12px]"
                        style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                      >
                        {r.reference_id ? r.reference_id.slice(0, 8) : '—'}
                      </span>
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
    </div>
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
    ['date', 'Date'],
    ['store', 'Store'],
    ['reason', 'Reason'],
    ['qty', 'Quantity'],
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
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ background: BLUE }}
                />
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
            {d === 'asc' ? 'Oldest first' : 'Newest first'}
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
