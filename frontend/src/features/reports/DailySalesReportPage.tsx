import { Link } from '@tanstack/react-router'
import { Download } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'
import { Label } from '@/shared/ui/label'

import { useDailySalesOrders, useOrgStores, useOrgTimezone } from './hooks'

const STATUS_STYLE: Record<string, string> = {
  open: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  completed: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  voided: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
  refunded: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  partially_refunded: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
}

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
  const timezone = useOrgTimezone()
  const stores = useOrgStores()

  const [date, setDate] = useState(() => todayInTimezone(timezone))
  const [storeId, setStoreId] = useState('')

  const range = useMemo(() => dayRangeUtc(date, timezone), [date, timezone])

  const { data: orders, isLoading } = useDailySalesOrders({
    dateFrom: range.from,
    dateTo: range.to,
    storeId,
  })

  const summary = useMemo(() => summarize(orders ?? []), [orders])

  const columns: DataTableColumn<OrderRow>[] = [
    {
      id: 'time',
      header: t('reports.time'),
      headerClassName: 'w-24',
      cell: (r) => formatTime(r.createdAt, timezone),
    },
    {
      id: 'order',
      header: t('reports.orderNumber'),
      cell: (r) => (
        <Link
          to="/daily-sales-report/$orderId"
          params={{ orderId: r.id }}
          className="font-medium hover:underline"
        >
          {r.orderNumber}
        </Link>
      ),
    },
    {
      id: 'cashier',
      header: t('reports.cashier'),
      cell: (r) => r.userName || '—',
    },
    {
      id: 'customer',
      header: t('reports.customer'),
      cell: (r) => r.customerName || '—',
    },
    {
      id: 'store',
      header: t('reports.store'),
      cell: (r) => r.storeName || '—',
    },
    {
      id: 'status',
      header: t('reports.status'),
      headerClassName: 'w-32',
      cell: (r) => (
        <span
          className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
            STATUS_STYLE[r.status] ?? STATUS_STYLE.open
          }`}
        >
          {t(`reports.orderStatus_${r.status}`, r.status)}
        </span>
      ),
    },
    {
      id: 'subtotal',
      header: t('reports.subtotal'),
      headerClassName: 'w-28 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.subtotal),
    },
    {
      id: 'tax',
      header: t('reports.tax'),
      headerClassName: 'w-24 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.taxTotal),
    },
    {
      id: 'discount',
      header: t('reports.discount'),
      headerClassName: 'w-24 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.discountTotal),
    },
    {
      id: 'total',
      header: t('reports.total'),
      headerClassName: 'w-28 text-right',
      className: 'text-right font-medium tabular-nums',
      cell: (r) => formatMoney(r.total),
    },
  ]

  const onExport = () => {
    downloadCsv(orders ?? [], date, timezone, t)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.dailySalesReport')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('reports.subtitle', { timezone })}
          </p>
        </div>
        <Button
          variant="outline"
          onClick={onExport}
          disabled={!orders || orders.length === 0}
        >
          <Download className="mr-2 h-4 w-4" />
          {t('reports.exportCsv')}
        </Button>
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <div className="space-y-1.5">
          <Label htmlFor="date">{t('reports.date')}</Label>
          <input
            id="date"
            type="date"
            className="h-10 rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </div>
        <div className="min-w-[14rem] space-y-1.5">
          <Label htmlFor="store">{t('reports.store')}</Label>
          <select
            id="store"
            className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
            value={storeId}
            onChange={(e) => setStoreId(e.target.value)}
          >
            <option value="">{t('reports.allStores')}</option>
            {stores.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <SummaryCard label={t('reports.orderCount')} value={String(summary.count)} />
        <SummaryCard label={t('reports.subtotal')} value={formatMoney(summary.subtotal)} />
        <SummaryCard label={t('reports.tax')} value={formatMoney(summary.tax)} />
        <SummaryCard label={t('reports.total')} value={formatMoney(summary.total)} emphasis />
      </div>

      <DataTable
        columns={columns}
        data={(orders ?? []) as OrderRow[]}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('reports.noOrders')}
      />
    </div>
  )
}

function SummaryCard({
  label,
  value,
  emphasis,
}: {
  label: string
  value: string
  emphasis?: boolean
}) {
  return (
    <div className="rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4">
      <div className="text-xs uppercase tracking-wide text-[color:var(--color-muted-foreground)]">
        {label}
      </div>
      <div
        className={`mt-1 text-xl tabular-nums ${
          emphasis ? 'font-semibold text-[color:var(--color-success)]' : 'font-medium'
        }`}
      >
        {value}
      </div>
    </div>
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

function formatTime(ts: OrderRow['createdAt'], timezone: string): string {
  if (!ts) return '—'
  const ms = Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)
  return new Date(ms).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: timezone,
  })
}

// todayInTimezone returns a Y-M-D string for "today" in the given IANA tz.
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

// dayRangeUtc returns [start, next-start) as UTC Dates for the given local Y-M-D in tz.
function dayRangeUtc(isoDate: string, timezone: string): { from: Date; to: Date } {
  const [y, m, d] = isoDate.split('-').map(Number)
  const from = zonedDateToUtc(y, m, d, timezone)
  const next = new Date(from.getTime() + 24 * 60 * 60 * 1000)
  // Handle DST transitions by re-anchoring next day at tz midnight.
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

// zonedDateToUtc: interpret Y-M-D 00:00 in tz as a UTC Date.
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

type TFn = (key: string, options?: Record<string, unknown>) => string

function downloadCsv(rows: OrderRow[], date: string, timezone: string, t: TFn): void {
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
