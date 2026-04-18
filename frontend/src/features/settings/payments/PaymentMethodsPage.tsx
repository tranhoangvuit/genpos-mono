import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { PaymentMethodDialog } from './PaymentMethodDialog'
import { useDeletePaymentMethod, usePaymentMethods } from './hooks'
import type { PaymentMethodRow } from './types'

export function PaymentMethodsPage() {
  const { t } = useTranslation()
  const { data: methods, isLoading } = usePaymentMethods()
  const deleteMut = useDeletePaymentMethod()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<PaymentMethodRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: PaymentMethodRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: PaymentMethodRow) => {
    if (!confirm(t('payments.confirmDelete', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<PaymentMethodRow>[] = [
    {
      id: 'name',
      header: t('payments.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'type',
      header: t('payments.type'),
      cell: (r) => t(`payments.type_${r.type}`, r.type),
      headerClassName: 'w-40',
    },
    {
      id: 'active',
      header: t('payments.isActive'),
      cell: (r) => (r.isActive ? t('common.yes') : t('common.no')),
      headerClassName: 'w-24',
    },
    {
      id: 'sortOrder',
      header: t('payments.sortOrder'),
      cell: (r) => r.sortOrder,
      headerClassName: 'w-24',
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
          <h1 className="text-2xl font-semibold">{t('nav.payments')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('payments.subtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('payments.newMethod')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={methods ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('payments.noMethods')}
      />

      <PaymentMethodDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}
