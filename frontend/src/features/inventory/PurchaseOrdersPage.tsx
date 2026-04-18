import { Link } from '@tanstack/react-router'
import { Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { usePurchaseOrders } from './hooks'
import type { PurchaseOrderListRow } from './types'

const STATUS_STYLE: Record<string, string> = {
  draft: 'bg-[color:var(--color-muted)] text-[color:var(--color-muted-foreground)]',
  submitted: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  partial: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  received: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  cancelled: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
}

export function PurchaseOrdersPage() {
  const { t } = useTranslation()
  const { data: orders, isLoading } = usePurchaseOrders()

  const columns: DataTableColumn<PurchaseOrderListRow>[] = [
    {
      id: 'poNumber',
      header: t('inventory.poNumber'),
      cell: (r) => (
        <Link
          to="/inventory/purchase-orders/$id"
          params={{ id: r.id }}
          className="font-medium hover:underline"
        >
          {r.po_number}
        </Link>
      ),
    },
    {
      id: 'supplier',
      header: t('inventory.supplier'),
      cell: (r) => r.supplier_name ?? '—',
    },
    {
      id: 'store',
      header: t('inventory.store'),
      cell: (r) => r.store_name ?? '—',
    },
    {
      id: 'status',
      header: t('inventory.status'),
      headerClassName: 'w-28',
      cell: (r) => (
        <span
          className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
            STATUS_STYLE[r.status] ?? STATUS_STYLE.draft
          }`}
        >
          {t(`inventory.status_${r.status}`, r.status)}
        </span>
      ),
    },
    {
      id: 'items',
      header: t('inventory.items'),
      headerClassName: 'w-20',
      cell: (r) => r.item_count,
    },
    {
      id: 'total',
      header: t('inventory.total'),
      headerClassName: 'w-32',
      cell: (r) => formatMoney(r.total),
    },
    {
      id: 'expected',
      header: t('inventory.expected'),
      headerClassName: 'w-32',
      cell: (r) => (r.expected_at ? formatDate(r.expected_at) : '—'),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.purchaseOrders')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('inventory.purchaseOrdersSubtitle')}
          </p>
        </div>
        <Button asChild>
          <Link to="/inventory/purchase-orders/new">
            <Plus className="mr-2 h-4 w-4" />
            {t('inventory.newPurchaseOrder')}
          </Link>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={orders ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('inventory.noPurchaseOrders')}
      />
    </div>
  )
}

function formatMoney(value: string): string {
  const n = Number(value)
  if (Number.isNaN(n)) return value
  return n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString()
}
