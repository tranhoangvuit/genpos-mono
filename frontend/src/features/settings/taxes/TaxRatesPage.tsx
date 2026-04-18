import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { TaxRateDialog } from './TaxRateDialog'
import { useDeleteTaxRate, useTaxRates } from './hooks'
import { fractionToPercent } from './schemas'
import type { TaxRateRow } from './types'

export function TaxRatesPage() {
  const { t } = useTranslation()
  const { data: rates, isLoading } = useTaxRates()
  const deleteMut = useDeleteTaxRate()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<TaxRateRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: TaxRateRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: TaxRateRow) => {
    if (!confirm(t('taxes.confirmDelete', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<TaxRateRow>[] = [
    {
      id: 'name',
      header: t('taxes.name'),
      cell: (r) => (
        <span className="font-medium">
          {r.name}
          {r.isDefault && (
            <span className="ml-2 rounded bg-[color:var(--color-muted)] px-1.5 py-0.5 text-xs">
              {t('taxes.defaultBadge')}
            </span>
          )}
        </span>
      ),
    },
    {
      id: 'rate',
      header: t('taxes.rate'),
      cell: (r) => `${fractionToPercent(r.rate)}%`,
      headerClassName: 'w-28',
    },
    {
      id: 'inclusive',
      header: t('taxes.isInclusive'),
      cell: (r) => (r.isInclusive ? t('common.yes') : t('common.no')),
      headerClassName: 'w-28',
    },
    {
      id: 'actions',
      header: '',
      headerClassName: 'w-24',
      cell: (r) => (
        <div className="flex justify-end gap-1">
          <Button variant="ghost" size="icon" onClick={() => onEdit(r)} aria-label={t('common.edit')}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onDelete(r)}
            disabled={deleteMut.isPending}
            aria-label={t('common.delete')}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.taxes')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('taxes.subtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('taxes.newRate')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={rates ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('taxes.noRates')}
      />

      <TaxRateDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}
