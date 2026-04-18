import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { StoreDialog } from './StoreDialog'
import { useDeleteStore, useStores } from './hooks'
import type { StoreRow } from './types'

export function StoresPage() {
  const { t } = useTranslation()
  const { data: stores, isLoading } = useStores()
  const deleteMut = useDeleteStore()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<StoreRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: StoreRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: StoreRow) => {
    if (!confirm(t('stores.confirmDelete', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<StoreRow>[] = [
    {
      id: 'name',
      header: t('stores.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'address',
      header: t('stores.address'),
      cell: (r) => r.address || '—',
    },
    {
      id: 'phone',
      header: t('stores.phone'),
      cell: (r) => r.phone || '—',
      headerClassName: 'w-40',
    },
    {
      id: 'email',
      header: t('stores.email'),
      cell: (r) => r.email || '—',
      headerClassName: 'w-56',
    },
    {
      id: 'status',
      header: t('stores.status'),
      cell: (r) => statusCell(t, r.status),
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
          <h1 className="text-2xl font-semibold">{t('nav.stores')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('stores.subtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('stores.newStore')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={stores ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('stores.noStores')}
      />

      <StoreDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}

function statusCell(t: (key: string) => string, status: string): string {
  if (status === 'active') return t('stores.statusActive')
  if (status === 'inactive') return t('stores.statusInactive')
  if (status === 'closed') return t('stores.statusClosed')
  return status || '—'
}
