import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { SupplierDialog } from './SupplierDialog'
import { useDeleteSupplier, useSuppliers } from './hooks'
import type { SupplierRow } from './types'

export function SuppliersPage() {
  const { t } = useTranslation()
  const { data: suppliers, isLoading } = useSuppliers()
  const deleteMut = useDeleteSupplier()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<SupplierRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: SupplierRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: SupplierRow) => {
    if (!confirm(t('inventory.confirmDeleteSupplier', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<SupplierRow>[] = [
    {
      id: 'name',
      header: t('inventory.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'contact',
      header: t('inventory.contactName'),
      cell: (r) => r.contact_name ?? '—',
    },
    {
      id: 'email',
      header: t('inventory.email'),
      cell: (r) => r.email ?? '—',
    },
    {
      id: 'phone',
      header: t('inventory.phone'),
      cell: (r) => r.phone ?? '—',
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
          <h1 className="text-2xl font-semibold">{t('nav.suppliers')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('inventory.suppliersSubtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('inventory.newSupplier')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={suppliers ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('inventory.noSuppliers')}
      />

      <SupplierDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}
