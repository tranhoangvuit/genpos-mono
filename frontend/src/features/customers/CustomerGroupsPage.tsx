import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { CustomerGroupDialog } from './CustomerGroupDialog'
import { useCustomerGroups, useDeleteCustomerGroup } from './hooks'
import type { CustomerGroupRow } from './types'

export function CustomerGroupsPage() {
  const { t } = useTranslation()
  const { data: groups, isLoading } = useCustomerGroups()
  const deleteMut = useDeleteCustomerGroup()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<CustomerGroupRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: CustomerGroupRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: CustomerGroupRow) => {
    if (!confirm(t('customers.confirmDeleteGroup', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<CustomerGroupRow>[] = [
    {
      id: 'name',
      header: t('customers.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'description',
      header: t('customers.description'),
      cell: (r) => r.description || '—',
    },
    {
      id: 'discount',
      header: t('customers.discount'),
      cell: (r) => formatDiscount(r.discountType, r.discountValue),
      headerClassName: 'w-32',
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
          <h1 className="text-2xl font-semibold">{t('nav.customerGroups')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('customers.groupsSubtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('customers.newGroup')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={groups ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('customers.noGroups')}
      />

      <CustomerGroupDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}

function formatDiscount(type: string, value: string): string {
  if (!type || !value) return '—'
  const n = Number(value)
  if (Number.isNaN(n)) return value
  if (type === 'percentage') return `${n}%`
  return n.toLocaleString()
}
