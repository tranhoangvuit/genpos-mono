import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { useOrder, useOrgTimezone } from './hooks'

type LineItem = {
  id: string
  productName: string
  variantName: string
  sku: string
  quantity: string
  unitPrice: string
  taxAmount: string
  discountAmount: string
  lineTotal: string
}

type Payment = {
  id: string
  paymentMethodName: string
  amount: string
  tendered: string
  changeAmount: string
  reference: string
  status: string
}

const STATUS_STYLE: Record<string, string> = {
  open: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  completed: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  voided: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
  refunded: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  partially_refunded: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
}

export function OrderDetailPage({ orderId }: { orderId: string }) {
  const { t } = useTranslation()
  const timezone = useOrgTimezone()
  const { data: order, isLoading, error } = useOrder(orderId)

  if (isLoading) {
    return <div className="p-6 text-[color:var(--color-muted-foreground)]">{t('common.loading')}</div>
  }
  if (error || !order) {
    return (
      <div className="space-y-4 p-6">
        <BackLink />
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {t('reports.orderNotFound')}
        </div>
      </div>
    )
  }

  const itemColumns: DataTableColumn<LineItem>[] = [
    {
      id: 'product',
      header: t('reports.product'),
      cell: (r) => (
        <div>
          <div className="font-medium">{r.productName}</div>
          <div className="text-xs text-[color:var(--color-muted-foreground)]">
            {r.variantName}
            {r.sku ? ` · ${r.sku}` : ''}
          </div>
        </div>
      ),
    },
    {
      id: 'quantity',
      header: t('reports.quantity'),
      headerClassName: 'w-24 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatQty(r.quantity),
    },
    {
      id: 'unitPrice',
      header: t('reports.unitPrice'),
      headerClassName: 'w-28 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.unitPrice),
    },
    {
      id: 'tax',
      header: t('reports.tax'),
      headerClassName: 'w-24 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.taxAmount),
    },
    {
      id: 'discount',
      header: t('reports.discount'),
      headerClassName: 'w-24 text-right',
      className: 'text-right tabular-nums',
      cell: (r) => formatMoney(r.discountAmount),
    },
    {
      id: 'total',
      header: t('reports.total'),
      headerClassName: 'w-28 text-right',
      className: 'text-right font-medium tabular-nums',
      cell: (r) => formatMoney(r.lineTotal),
    },
  ]

  const paymentColumns: DataTableColumn<Payment>[] = [
    { id: 'method', header: t('reports.paymentMethod'), cell: (p) => p.paymentMethodName || '—' },
    {
      id: 'amount',
      header: t('reports.amount'),
      headerClassName: 'w-28 text-right',
      className: 'text-right tabular-nums',
      cell: (p) => formatMoney(p.amount),
    },
    {
      id: 'tendered',
      header: t('reports.tendered'),
      headerClassName: 'w-28 text-right',
      className: 'text-right tabular-nums',
      cell: (p) => (p.tendered ? formatMoney(p.tendered) : '—'),
    },
    {
      id: 'change',
      header: t('reports.change'),
      headerClassName: 'w-28 text-right',
      className: 'text-right tabular-nums',
      cell: (p) => (p.changeAmount ? formatMoney(p.changeAmount) : '—'),
    },
    { id: 'reference', header: t('reports.reference'), cell: (p) => p.reference || '—' },
    { id: 'status', header: t('reports.status'), cell: (p) => p.status },
  ]

  return (
    <div className="space-y-6">
      <BackLink />

      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">
            {t('reports.orderNumber')} {order.orderNumber}
          </h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {formatDateTime(order.createdAt, timezone)}
            {order.storeName ? ` · ${order.storeName}` : ''}
          </p>
        </div>
        <span
          className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
            STATUS_STYLE[order.status] ?? STATUS_STYLE.open
          }`}
        >
          {t(`reports.orderStatus_${order.status}`, order.status)}
        </span>
      </div>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <InfoTile label={t('reports.cashier')} value={order.userName || '—'} />
        <InfoTile label={t('reports.customer')} value={order.customerName || '—'} />
        <InfoTile label={t('reports.subtotal')} value={formatMoney(order.subtotal)} />
        <InfoTile label={t('reports.tax')} value={formatMoney(order.taxTotal)} />
        <InfoTile label={t('reports.discount')} value={formatMoney(order.discountTotal)} />
        <InfoTile
          label={t('reports.total')}
          value={formatMoney(order.total)}
          emphasis
        />
      </div>

      <section className="space-y-2">
        <h2 className="text-lg font-semibold">{t('reports.lineItems')}</h2>
        <DataTable
          columns={itemColumns}
          data={order.lineItems as LineItem[]}
          rowKey={(r) => r.id}
          emptyMessage={t('reports.noLineItems')}
        />
      </section>

      <section className="space-y-2">
        <h2 className="text-lg font-semibold">{t('reports.payments')}</h2>
        <DataTable
          columns={paymentColumns}
          data={order.payments as Payment[]}
          rowKey={(p) => p.id}
          emptyMessage={t('reports.noPayments')}
        />
      </section>

      {order.notes && (
        <section className="space-y-2">
          <h2 className="text-lg font-semibold">{t('reports.notes')}</h2>
          <p className="rounded-md border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-3 text-sm">
            {order.notes}
          </p>
        </section>
      )}
    </div>
  )
}

function BackLink() {
  const { t } = useTranslation()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  return (
    <Link
      to="/$subdomain/daily-sales-report"
      params={{ subdomain }}
      className="inline-flex items-center text-sm text-[color:var(--color-muted-foreground)] hover:text-[color:var(--color-foreground)]"
    >
      <ArrowLeft className="mr-1 h-4 w-4" />
      {t('reports.backToReport')}
    </Link>
  )
}

function InfoTile({
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

function formatMoney(value: string): string {
  const n = Number(value)
  if (Number.isNaN(n)) return value
  return n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

function formatQty(value: string): string {
  const n = Number(value)
  if (Number.isNaN(n)) return value
  return n.toLocaleString(undefined, { maximumFractionDigits: 4 })
}

function formatDateTime(
  ts: { seconds: bigint; nanos: number } | undefined,
  timezone: string,
): string {
  if (!ts) return '—'
  const ms = Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)
  return new Date(ms).toLocaleString(undefined, { timeZone: timezone })
}
