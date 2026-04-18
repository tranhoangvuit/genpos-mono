import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2, Upload } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'

import { CustomerDialog } from './CustomerDialog'
import { ImportCustomerDialog } from './ImportCustomerDialog'
import { useCustomers, useDeleteCustomer, useGetCustomer } from './hooks'
import type { CustomerListRow, CustomerRow } from './types'

export function CustomersPage() {
  const { t } = useTranslation()
  const { data: customers, isLoading } = useCustomers()
  const deleteMut = useDeleteCustomer()
  const getMut = useGetCustomer()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<CustomerRow | null>(null)
  const [query, setQuery] = useState('')
  const [importOpen, setImportOpen] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const needle = query.trim().toLowerCase()
  const rows = (customers ?? []).filter((r) =>
    needle === ''
      ? true
      : r.name.toLowerCase().includes(needle) ||
        r.email.toLowerCase().includes(needle) ||
        r.phone.toLowerCase().includes(needle),
  )

  const onEdit = async (r: CustomerListRow) => {
    const res = await getMut.mutateAsync(r.id)
    if (res.customer) {
      setEditing(res.customer)
      setDialogOpen(true)
    }
  }

  const onDelete = async (r: CustomerListRow) => {
    if (!confirm(t('customers.confirmDelete', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<CustomerListRow>[] = [
    {
      id: 'name',
      header: t('customers.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'email',
      header: t('customers.email'),
      cell: (r) => r.email || '—',
    },
    {
      id: 'phone',
      header: t('customers.phone'),
      cell: (r) => r.phone || '—',
    },
    {
      id: 'groups',
      header: t('customers.groups'),
      cell: (r) => r.groupNames || '—',
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
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.customers')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('customers.subtitle')}
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setImportOpen(true)}>
            <Upload className="mr-2 h-4 w-4" />
            {t('customers.importCustomers')}
          </Button>
          <Button
            onClick={() => {
              setEditing(null)
              setDialogOpen(true)
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            {t('customers.newCustomer')}
          </Button>
        </div>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <div className="max-w-sm">
        <Input
          placeholder={t('customers.searchPlaceholder')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={rows}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('customers.noCustomers')}
      />

      <CustomerDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
      <ImportCustomerDialog open={importOpen} onOpenChange={setImportOpen} />
    </div>
  )
}
