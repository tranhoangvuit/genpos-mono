import { Link } from '@tanstack/react-router'
import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  Calendar,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Download,
  Filter,
  HelpCircle,
  Receipt,
  Search,
  Store,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'

import { useAuthStore } from '@/shared/auth/store'

import { useDailySalesOrders, useOrgStores, useOrgTimezone } from './hooks'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE = 'hsl(142.1 76.2% 36.3%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const IDLE_SOFT = 'hsl(210 40% 96.1%)'
const IDLE_INK = 'hsl(215.3 25% 26.7%)'
const WARN_SOFT = 'hsl(48 96% 94%)'
const WARN_INK = 'hsl(31 92% 28%)'
const WARN = 'hsl(38 92% 50%)'
const DESTRUCT_SOFT = 'hsl(0 93% 94%)'
const DESTRUCT_INK = 'hsl(0 74% 42%)'
const DESTRUCT = 'hsl(0 84% 60%)'
const BLUE_SOFT = 'hsl(214 100% 96.7%)'

type StatusKey = 'all' | 'completed' | 'open' | 'voided' | 'refunded'
type SortDir = 'asc' | 'desc'
type SortKey = 'time' | 'total' | 'order'

const PAGE_SIZE = 15

type OrderRow = {
  id: string
  orderNumber: string
  status: string
  subtotal: string
  taxTotal: string
  discountTotal: string
  total: string
  storeId: string
  storeName: string
  userName: string
  customerName: string
  createdAt?: { seconds: bigint; nanos: number }
}

export function DailySalesReportPage() {
  const { t } = useTranslation()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const timezone = useOrgTimezone()
  const stores = useOrgStores()

  const [date, setDate] = useState(() => todayInTimezone(timezone))
  const [storeId, setStoreId] = useState('')
  const [query, setQuery] = useState('')
  const [tab, setTab] = useState<StatusKey>('all')
  const [sortKey, setSortKey] = useState<SortKey>('time')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [page, setPage] = useState(0)
  const sortRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
    }
    if (sortOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen])

  const range = useMemo(() => dayRangeUtc(date, timezone), [date, timezone])

  const { data: orders, isLoading } = useDailySalesOrders({
    dateFrom: range.from,
    dateTo: range.to,
    storeId,
  })

  const all = (orders ?? []) as OrderRow[]

  const counts = useMemo(
    () => ({
      all: all.length,
      completed: all.filter((o) => o.status === 'completed').length,
      open: all.filter((o) => o.status === 'open').length,
      voided: all.filter((o) => o.status === 'voided').length,
      refunded: all.filter(
        (o) => o.status === 'refunded' || o.status === 'partially_refunded',
      ).length,
    }),
    [all],
  )

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let list = all
    if (tab !== 'all') {
      list = list.filter((o) => {
        if (tab === 'refunded')
          return o.status === 'refunded' || o.status === 'partially_refunded'
        return o.status === tab
      })
    }
    if (q) {
      list = list.filter(
        (o) =>
          o.orderNumber.toLowerCase().includes(q) ||
          (o.customerName || '').toLowerCase().includes(q) ||
          (o.userName || '').toLowerCase().includes(q),
      )
    }
    list = [...list].sort((a, b) => {
      let cmp = 0
      if (sortKey === 'time') {
        cmp = tsMs(a.createdAt) - tsMs(b.createdAt)
      } else if (sortKey === 'total') {
        cmp = Number(a.total) - Number(b.total)
      } else {
        cmp = a.orderNumber.localeCompare(b.orderNumber)
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return list
  }, [all, query, tab, sortKey, sortDir])

  const summary = useMemo(() => summarize(filtered), [filtered])

  const pageStart = page * PAGE_SIZE
  const pageRows = filtered.slice(pageStart, pageStart + PAGE_SIZE)
  const pageEnd = pageStart + pageRows.length

  const selectedStore = stores.find((s) => s.id === storeId)

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <Receipt className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('nav.dailySalesReport')}
          <span
            className="rounded-md px-2 py-0.5 text-[13px] font-medium"
            style={{ background: MUTED, color: MUTED_FG }}
          >
            {all.length}
          </span>
        </h1>
        <div className="flex gap-2">
          <MoreBtn
            icon={<Download className="h-3.5 w-3.5" />}
            onClick={() => downloadCsv(filtered, date, timezone, t)}
            disabled={filtered.length === 0}
          >
            {t('reports.exportCsv')}
          </MoreBtn>
        </div>
      </div>

      <section
        className="mb-4 grid grid-cols-[auto_auto_repeat(4,1fr)] rounded-lg border bg-white p-0.5"
        style={{ borderColor: BORDER }}
      >
        <KpiCell divider={false}>
          <label
            className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-white px-2.5 py-1.5 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
            style={{ borderColor: BORDER }}
          >
            <Calendar className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
            <input
              type="date"
              value={date}
              onChange={(e) => {
                setDate(e.target.value)
                setPage(0)
              }}
              className="cursor-pointer border-none bg-transparent p-0 text-[12.5px] font-medium outline-none"
              style={{ color: FG }}
            />
          </label>
        </KpiCell>
        <KpiCell>
          <label
            className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-white px-2.5 py-1.5 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
            style={{ borderColor: BORDER }}
          >
            <Store className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
            <select
              value={storeId}
              onChange={(e) => {
                setStoreId(e.target.value)
                setPage(0)
              }}
              className="cursor-pointer border-none bg-transparent p-0 pr-4 text-[12.5px] font-medium outline-none"
              style={{ color: selectedStore ? FG : MUTED_FG }}
            >
              <option value="">{t('reports.allStores')}</option>
              {stores.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </label>
        </KpiCell>
        <KpiCell>
          <KpiLabel>{t('reports.orderCount')}</KpiLabel>
          <div className="mt-1 text-[20px] font-semibold tabular-nums">
            {summary.count}
          </div>
        </KpiCell>
        <KpiCell>
          <KpiLabel>{t('reports.subtotal')}</KpiLabel>
          <div className="mt-1 text-[20px] font-semibold tabular-nums">
            {formatMoney(summary.subtotal)}
          </div>
        </KpiCell>
        <KpiCell>
          <KpiLabel>{t('reports.tax')}</KpiLabel>
          <div className="mt-1 text-[20px] font-semibold tabular-nums">
            {formatMoney(summary.tax)}
          </div>
        </KpiCell>
        <KpiCell>
          <KpiLabel>{t('reports.total')}</KpiLabel>
          <div className="mt-1 flex items-baseline gap-1.5 text-[20px] font-semibold tabular-nums">
            <span
              className="underline decoration-2 underline-offset-[3px]"
              style={{ textDecorationColor: BLUE }}
            >
              {formatMoney(summary.total)}
            </span>
          </div>
        </KpiCell>
      </section>

      <section
        className="overflow-hidden rounded-lg border bg-white"
        style={{ borderColor: BORDER }}
      >
        <div
          className="flex items-center justify-between border-b px-2.5 pt-1.5"
          style={{ borderColor: BORDER }}
        >
          <div className="flex items-center gap-0.5">
            {(['all', 'completed', 'open', 'voided', 'refunded'] as StatusKey[]).map(
              (k) => (
                <button
                  key={k}
                  type="button"
                  onClick={() => {
                    setTab(k)
                    setPage(0)
                  }}
                  className="-mb-px inline-flex cursor-pointer items-center gap-1.5 rounded-t-md px-3 pb-2.5 pt-2 text-[13px] transition"
                  style={{
                    color: tab === k ? FG : MUTED_FG,
                    fontWeight: tab === k ? 600 : 500,
                    borderBottom: `2px solid ${tab === k ? FG : 'transparent'}`,
                  }}
                >
                  {tabLabel(k, t)}
                  <span className="text-[11.5px]" style={{ color: MUTED_FG }}>
                    {counts[k]}
                  </span>
                </button>
              ),
            )}
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
                  placeholder={t('reports.searchPlaceholder', 'Search orders, cashiers, customers')}
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
            </div>
          </div>
        )}

        <div className="w-full overflow-x-auto">
          <table className="w-full border-collapse" style={{ tableLayout: 'fixed' }}>
            <colgroup>
              <col style={{ width: 80 }} />
              <col style={{ width: 140 }} />
              <col style={{ width: 160 }} />
              <col style={{ width: 180 }} />
              <col style={{ width: 160 }} />
              <col style={{ width: 140 }} />
              <col style={{ width: 120 }} />
              <col style={{ width: 100 }} />
              <col style={{ width: 110 }} />
              <col style={{ width: 120 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>{t('reports.time')}</Th>
                <Th>{t('reports.orderNumber')}</Th>
                <Th>{t('reports.cashier')}</Th>
                <Th>{t('reports.customer')}</Th>
                <Th>{t('reports.store')}</Th>
                <Th>{t('reports.status')}</Th>
                <Th align="right">{t('reports.subtotal')}</Th>
                <Th align="right">{t('reports.tax')}</Th>
                <Th align="right">{t('reports.discount')}</Th>
                <Th align="right">{t('reports.total')}</Th>
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td
                    colSpan={10}
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
                    colSpan={10}
                    className="py-10 text-center text-[13px]"
                    style={{ color: MUTED_FG }}
                  >
                    {t('reports.noOrders')}
                  </td>
                </tr>
              )}
              {pageRows.map((o) => (
                <tr
                  key={o.id}
                  className="transition hover:bg-[hsl(210_40%_96%_/_0.3)]"
                >
                  <Td>{formatTime(o.createdAt, timezone)}</Td>
                  <Td>
                    <Link
                      to="/$subdomain/daily-sales-report/$orderId"
                      params={{ subdomain, orderId: o.id }}
                      className="font-medium hover:underline"
                      style={{ color: FG, textDecorationColor: MUTED_FG }}
                    >
                      {o.orderNumber}
                    </Link>
                  </Td>
                  <Td>{o.userName || '—'}</Td>
                  <Td>{o.customerName || '—'}</Td>
                  <Td>
                    {o.storeName ? (
                      <span
                        className="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[12px]"
                        style={{ background: MUTED, color: FG }}
                      >
                        {o.storeName}
                      </span>
                    ) : (
                      <span style={{ color: MUTED_FG }}>—</span>
                    )}
                  </Td>
                  <Td>
                    <StatusPill status={o.status} t={t} />
                  </Td>
                  <Td align="right">
                    <span className="tabular-nums">{formatMoney(o.subtotal)}</span>
                  </Td>
                  <Td align="right">
                    <span className="tabular-nums">{formatMoney(o.taxTotal)}</span>
                  </Td>
                  <Td align="right">
                    <span className="tabular-nums">
                      {formatMoney(o.discountTotal)}
                    </span>
                  </Td>
                  <Td align="right">
                    <span className="font-medium tabular-nums">
                      {formatMoney(o.total)}
                    </span>
                  </Td>
                </tr>
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
    </div>
  )
}

function tabLabel(k: StatusKey, t: TFunction): string {
  if (k === 'all') return 'All'
  if (k === 'completed') return t('reports.orderStatus_completed', 'Completed')
  if (k === 'open') return t('reports.orderStatus_open', 'Open')
  if (k === 'voided') return t('reports.orderStatus_voided', 'Voided')
  return t('reports.orderStatus_refunded', 'Refunded')
}

function KpiCell({
  children,
  divider = true,
}: {
  children: React.ReactNode
  divider?: boolean
}) {
  return (
    <div
      className="min-w-0 px-4 py-3.5"
      style={{ borderLeft: divider ? `1px solid ${BORDER}` : undefined }}
    >
      {children}
    </div>
  )
}

function KpiLabel({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="inline-flex items-center gap-1 text-[12.5px] font-medium"
      style={{ color: MUTED_FG }}
    >
      {children}
      <HelpCircle className="h-3 w-3" strokeWidth={2} />
    </div>
  )
}

function MoreBtn({
  children,
  icon,
  onClick,
  disabled,
}: {
  children: React.ReactNode
  icon?: React.ReactNode
  onClick?: () => void
  disabled?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="inline-flex h-9 cursor-pointer items-center gap-1.5 rounded-md border bg-white px-3.5 text-[13px] font-medium transition hover:bg-[hsl(210_40%_96%)] disabled:cursor-not-allowed disabled:opacity-50"
      style={{ borderColor: BORDER }}
    >
      {icon}
      {children}
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
    ['time', 'Time'],
    ['order', 'Order number'],
    ['total', 'Total'],
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
      className="px-3 py-3 align-middle text-[13px]"
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

function StatusPill({ status, t }: { status: string; t: TFunction }) {
  const palette = statusPalette(status)
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2 py-[2px] text-[11.5px] font-medium"
      style={{ background: palette.bg, color: palette.color }}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: palette.dot }}
      />
      {t(`reports.orderStatus_${status}`, status)}
    </span>
  )
}

function statusPalette(status: string): { bg: string; color: string; dot: string } {
  switch (status) {
    case 'completed':
      return { bg: DONE_SOFT, color: DONE_INK, dot: DONE }
    case 'voided':
      return { bg: DESTRUCT_SOFT, color: DESTRUCT_INK, dot: DESTRUCT }
    case 'refunded':
    case 'partially_refunded':
      return { bg: WARN_SOFT, color: WARN_INK, dot: WARN }
    case 'open':
      return { bg: BLUE_SOFT, color: BLUE_INK, dot: BLUE }
    default:
      return { bg: IDLE_SOFT, color: IDLE_INK, dot: MUTED_FG }
  }
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

function summarize(orders: OrderRow[]): {
  count: number
  subtotal: string
  tax: string
  total: string
} {
  let subtotal = 0
  let tax = 0
  let total = 0
  for (const o of orders) {
    subtotal += Number(o.subtotal) || 0
    tax += Number(o.taxTotal) || 0
    total += Number(o.total) || 0
  }
  return {
    count: orders.length,
    subtotal: subtotal.toFixed(2),
    tax: tax.toFixed(2),
    total: total.toFixed(2),
  }
}

function formatMoney(value: string): string {
  const n = Number(value)
  if (Number.isNaN(n)) return value
  return n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

function tsMs(ts: OrderRow['createdAt']): number {
  if (!ts) return 0
  return Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)
}

function formatTime(ts: OrderRow['createdAt'], timezone: string): string {
  if (!ts) return '—'
  const ms = tsMs(ts)
  return new Date(ms).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: timezone,
  })
}

function todayInTimezone(timezone: string): string {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(new Date())
  const y = parts.find((p) => p.type === 'year')?.value ?? '1970'
  const m = parts.find((p) => p.type === 'month')?.value ?? '01'
  const d = parts.find((p) => p.type === 'day')?.value ?? '01'
  return `${y}-${m}-${d}`
}

function dayRangeUtc(isoDate: string, timezone: string): { from: Date; to: Date } {
  const [y, m, d] = isoDate.split('-').map(Number)
  const from = zonedDateToUtc(y, m, d, timezone)
  const next = new Date(from.getTime() + 24 * 60 * 60 * 1000)
  const nextParts = new Intl.DateTimeFormat('en-CA', {
    timeZone: timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(next)
  const ny = Number(nextParts.find((p) => p.type === 'year')?.value)
  const nm = Number(nextParts.find((p) => p.type === 'month')?.value)
  const nd = Number(nextParts.find((p) => p.type === 'day')?.value)
  const to = zonedDateToUtc(ny, nm, nd, timezone)
  return { from, to }
}

function zonedDateToUtc(y: number, m: number, d: number, timezone: string): Date {
  const guess = Date.UTC(y, m - 1, d)
  const offset1 = tzOffsetMs(new Date(guess), timezone)
  const adjusted = new Date(guess - offset1)
  const offset2 = tzOffsetMs(adjusted, timezone)
  return offset1 === offset2 ? adjusted : new Date(guess - offset2)
}

function tzOffsetMs(utcDate: Date, timezone: string): number {
  const tzString = utcDate.toLocaleString('en-US', { timeZone: timezone })
  const utcString = utcDate.toLocaleString('en-US', { timeZone: 'UTC' })
  return new Date(tzString).getTime() - new Date(utcString).getTime()
}

function downloadCsv(rows: OrderRow[], date: string, timezone: string, t: TFunction): void {
  const headers = [
    t('reports.time'),
    t('reports.orderNumber'),
    t('reports.cashier'),
    t('reports.customer'),
    t('reports.store'),
    t('reports.status'),
    t('reports.subtotal'),
    t('reports.tax'),
    t('reports.discount'),
    t('reports.total'),
  ]
  const lines = [headers.map(csvField).join(',')]
  for (const r of rows) {
    lines.push(
      [
        formatTime(r.createdAt, timezone),
        r.orderNumber,
        r.userName,
        r.customerName,
        r.storeName,
        r.status,
        r.subtotal,
        r.taxTotal,
        r.discountTotal,
        r.total,
      ]
        .map(csvField)
        .join(','),
    )
  }
  const blob = new Blob(['\uFEFF' + lines.join('\r\n')], {
    type: 'text/csv;charset=utf-8',
  })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `daily-sales-${date}.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function csvField(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}
