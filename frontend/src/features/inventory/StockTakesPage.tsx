import { timestampDate } from '@bufbuild/protobuf/wkt'
import { ConnectError } from '@connectrpc/connect'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ClipboardList,
  Filter,
  Plus,
  Search,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

import { useCreateStockTake, useStockTakes, useStores } from './hooks'
import type { StockTakeListRow } from './types'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const BLUE_SOFT = 'hsl(214 100% 96.7%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const IDLE_SOFT = 'hsl(210 40% 96.1%)'
const IDLE_INK = 'hsl(215.3 25% 26.7%)'

const PAGE_SIZE = 15

type TabKey = 'all' | 'in_progress' | 'completed' | 'cancelled'
type SortKey = 'date' | 'store' | 'status' | 'items' | 'variance' | 'completed'
type SortDir = 'asc' | 'desc'

function statusPalette(s: string) {
  switch (s) {
    case 'in_progress':
      return { bg: BLUE_SOFT, color: BLUE_INK }
    case 'completed':
      return { bg: DONE_SOFT, color: DONE_INK }
    case 'cancelled':
      return { bg: IDLE_SOFT, color: IDLE_INK }
    default:
      return { bg: MUTED, color: MUTED_FG }
  }
}

function formatDate(ts: StockTakeListRow['createdAt']) {
  if (!ts) return '—'
  try {
    return timestampDate(ts).toLocaleDateString('en-US', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    })
  } catch {
    return '—'
  }
}

export function StockTakesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const { data: takes, isLoading } = useStockTakes()
  const { data: stores } = useStores()
  const create = useCreateStockTake()

  const [tab, setTab] = useState<TabKey>('all')
  const [query, setQuery] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [sortKey, setSortKey] = useState<SortKey>('date')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const sortRef = useRef<HTMLDivElement>(null)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [storeId, setStoreId] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
    }
    if (sortOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen])

  const all = takes ?? []

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let arr = all
    if (tab !== 'all') arr = arr.filter((r) => r.status === tab)
    if (q) {
      arr = arr.filter((r) => r.storeName.toLowerCase().includes(q))
    }
    arr = [...arr].sort((a, b) => {
      let cmp = 0
      switch (sortKey) {
        case 'store':
          cmp = a.storeName.localeCompare(b.storeName)
          break
        case 'status':
          cmp = a.status.localeCompare(b.status)
          break
        case 'items':
          cmp = a.itemCount - b.itemCount
          break
        case 'variance':
          cmp = a.varianceLines - b.varianceLines
          break
        case 'completed': {
          const ax = a.completedAt ? Number(a.completedAt.seconds) : 0
          const bx = b.completedAt ? Number(b.completedAt.seconds) : 0
          cmp = ax - bx
          break
        }
        default: {
          const ax = a.createdAt ? Number(a.createdAt.seconds) : 0
          const bx = b.createdAt ? Number(b.createdAt.seconds) : 0
          cmp = ax - bx
        }
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return arr
  }, [all, tab, query, sortKey, sortDir])

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

  const openDialog = () => {
    setStoreId(stores?.[0]?.id ?? '')
    setNotes('')
    setError(null)
    setDialogOpen(true)
    create.reset()
  }

  const onCreate = async () => {
    setError(null)
    if (!storeId) {
      setError(t('inventory.validation.storeRequired'))
      return
    }
    try {
      const res = await create.mutateAsync({ storeId, notes })
      setDialogOpen(false)
      if (res.stockTake?.id) {
        void navigate({
          to: '/$subdomain/inventory/stock-takes/$id',
          params: { subdomain, id: res.stockTake.id },
        })
      }
    } catch (err) {
      setError(ConnectError.from(err).rawMessage)
    }
  }

  const TAB_LABEL: Record<TabKey, string> = {
    all: 'All',
    in_progress: t('inventory.takeStatus_in_progress'),
    completed: t('inventory.takeStatus_completed'),
    cancelled: t('inventory.takeStatus_cancelled'),
  }

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <ClipboardList className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('nav.stockTakes')}
        </h1>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={openDialog}
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
            {t('inventory.newStockTake')}
          </button>
        </div>
      </div>

      <section
        className="overflow-hidden rounded-lg border bg-white"
        style={{ borderColor: BORDER }}
      >
        <div
          className="flex items-center justify-between border-b px-2.5 pt-1.5"
          style={{ borderColor: BORDER }}
        >
          <div className="flex items-center gap-0.5">
            {(['all', 'in_progress', 'completed', 'cancelled'] as TabKey[]).map((k) => (
              <button
                key={k}
                type="button"
                onClick={() => {
                  setTab(k)
                  setPage(0)
                }}
                className="-mb-px inline-flex cursor-pointer items-center rounded-t-md px-3 pb-2.5 pt-2 text-[13px] transition"
                style={{
                  color: tab === k ? FG : MUTED_FG,
                  fontWeight: tab === k ? 600 : 500,
                  borderBottom: `2px solid ${tab === k ? FG : 'transparent'}`,
                }}
              >
                {TAB_LABEL[k]}
              </button>
            ))}
            <button
              type="button"
              aria-label="Add view"
              className="inline-flex cursor-pointer items-center rounded-md px-2 py-1.5 transition hover:bg-[hsl(210_40%_96%)]"
              style={{ color: MUTED_FG }}
            >
              <Plus className="h-3.5 w-3.5" strokeWidth={2.2} />
            </button>
          </div>
          <div className="flex items-center gap-1 pb-1.5">
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
                  placeholder="Search stock takes"
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
                {t('common.cancel')}
              </button>
              <button
                type="button"
                className="cursor-default rounded-md px-2.5 py-2 text-[13px] font-medium"
                style={{ color: MUTED_FG }}
              >
                Save as
              </button>
            </div>
            <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
              <FilterChip>Stores</FilterChip>
              <FilterChip>Statuses</FilterChip>
              <button
                type="button"
                className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
                style={{ borderColor: BORDER, color: MUTED_FG }}
              >
                <Plus className="h-3 w-3" strokeWidth={2.2} />
                Add filter
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
              <col style={{ width: 120 }} />
              <col style={{ width: 90 }} />
              <col style={{ width: 120 }} />
              <col style={{ width: 140 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th>{t('inventory.stockTake')}</Th>
                <Th>{t('inventory.store')}</Th>
                <Th>{t('inventory.status')}</Th>
                <Th align="right">{t('inventory.items')}</Th>
                <Th align="right">{t('inventory.varianceLines')}</Th>
                <Th>{t('inventory.completedAt')}</Th>
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
                    {t('common.loading')}
                  </td>
                </tr>
              )}
              {!isLoading && pageRows.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="py-10 text-center text-[13px]"
                    style={{ color: MUTED_FG }}
                  >
                    {t('inventory.noStockTakes')}
                  </td>
                </tr>
              )}
              {pageRows.map((r) => (
                <Row
                  key={r.id}
                  r={r}
                  subdomain={subdomain}
                  checked={selected.has(r.id)}
                  onToggle={() => toggle(r.id)}
                  statusText={t(`inventory.takeStatus_${r.status}`, r.status)}
                />
              ))}
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

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('inventory.newStockTake')}</DialogTitle>
            <DialogDescription>{t('inventory.newStockTakeSubtitle')}</DialogDescription>
          </DialogHeader>

          {error && (
            <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="storeId">{t('inventory.store')}</Label>
              <select
                id="storeId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                value={storeId}
                onChange={(e) => setStoreId(e.target.value)}
              >
                <option value="">—</option>
                {(stores ?? []).map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="notes">{t('inventory.notes')}</Label>
              <Textarea
                id="notes"
                rows={3}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={create.isPending}
            >
              {t('common.cancel')}
            </Button>
            <Button type="button" onClick={onCreate} disabled={create.isPending}>
              {create.isPending ? t('common.loading') : t('inventory.startCount')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function Row({
  r,
  subdomain,
  checked,
  onToggle,
  statusText,
}: {
  r: StockTakeListRow
  subdomain: string
  checked: boolean
  onToggle: () => void
  statusText: string
}) {
  const pal = statusPalette(r.status)
  return (
    <tr className="transition hover:bg-[hsl(210_40%_96%_/_0.3)]">
      <Td>
        <Check checked={checked} onClick={onToggle} />
      </Td>
      <Td>
        <Link
          to="/$subdomain/inventory/stock-takes/$id"
          params={{ subdomain, id: r.id }}
          className="font-semibold hover:underline"
          style={{ color: FG }}
        >
          {formatDate(r.createdAt)}
        </Link>
      </Td>
      <Td>{r.storeName || '—'}</Td>
      <Td>
        <span
          className="inline-flex items-center rounded-full px-2.5 py-[2px] text-[12px] font-medium"
          style={{ background: pal.bg, color: pal.color }}
        >
          {statusText}
        </span>
      </Td>
      <Td align="right">
        <span className="tabular-nums">{r.itemCount}</span>
      </Td>
      <Td align="right">
        <span className="tabular-nums">{r.varianceLines}</span>
      </Td>
      <Td>
        <span style={{ color: MUTED_FG }}>{formatDate(r.completedAt)}</span>
      </Td>
    </tr>
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
    ['date', 'Created'],
    ['store', 'Store'],
    ['status', 'Status'],
    ['items', 'Items'],
    ['variance', 'Variance lines'],
    ['completed', 'Completed'],
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

function FilterChip({ children }: { children: React.ReactNode }) {
  return (
    <button
      type="button"
      className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
      style={{ background: MUTED, borderColor: BORDER, color: FG }}
    >
      {children}
      <ChevronDown className="h-3 w-3" style={{ color: MUTED_FG }} />
    </button>
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
