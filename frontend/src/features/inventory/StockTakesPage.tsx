import { ConnectError } from '@connectrpc/connect'
import { Link, useNavigate } from '@tanstack/react-router'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'
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

import {
  useCreateStockTake,
  useStockTakes,
  useStores,
} from './hooks'
import type { StockTakeListRow } from './types'

const STATUS_STYLE: Record<string, string> = {
  in_progress: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  completed: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  cancelled: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
}

export function StockTakesPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: takes, isLoading } = useStockTakes()
  const { data: stores } = useStores()
  const create = useCreateStockTake()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [storeId, setStoreId] = useState('')
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)

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
          to: '/inventory/stock-takes/$id',
          params: { id: res.stockTake.id },
        })
      }
    } catch (err) {
      setError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<StockTakeListRow>[] = [
    {
      id: 'id',
      header: t('inventory.stockTake'),
      cell: (r) => (
        <Link
          to="/inventory/stock-takes/$id"
          params={{ id: r.id }}
          className="font-medium hover:underline"
        >
          {formatDate(r.created_at)}
        </Link>
      ),
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
            STATUS_STYLE[r.status] ?? STATUS_STYLE.in_progress
          }`}
        >
          {t(`inventory.takeStatus_${r.status}`, r.status)}
        </span>
      ),
    },
    {
      id: 'items',
      header: t('inventory.items'),
      headerClassName: 'w-24',
      cell: (r) => r.item_count,
    },
    {
      id: 'variance',
      header: t('inventory.varianceLines'),
      headerClassName: 'w-32',
      cell: (r) => r.variance_lines ?? 0,
    },
    {
      id: 'completed',
      header: t('inventory.completedAt'),
      headerClassName: 'w-40',
      cell: (r) => (r.completed_at ? formatDate(r.completed_at) : '—'),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.stockTakes')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('inventory.stockTakesSubtitle')}
          </p>
        </div>
        <Button onClick={openDialog}>
          <Plus className="mr-2 h-4 w-4" />
          {t('inventory.newStockTake')}
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={takes ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('inventory.noStockTakes')}
      />

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

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}
