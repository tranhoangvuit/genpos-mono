import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'

import { MemberDialog } from './MemberDialog'
import { useDeleteMember, useMembers } from './hooks'
import type { MemberRow } from './types'

export function MembersPage() {
  const { t } = useTranslation()
  const { data: members, isLoading } = useMembers()
  const deleteMut = useDeleteMember()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<MemberRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const onEdit = (r: MemberRow) => {
    setEditing(r)
    setDialogOpen(true)
  }

  const onDelete = async (r: MemberRow) => {
    if (!confirm(t('members.confirmDelete', { name: r.name }))) return
    setDeleteError(null)
    try {
      await deleteMut.mutateAsync(r.id)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
    }
  }

  const columns: DataTableColumn<MemberRow>[] = [
    {
      id: 'name',
      header: t('members.name'),
      cell: (r) => <span className="font-medium">{r.name}</span>,
    },
    {
      id: 'email',
      header: t('members.email'),
      cell: (r) => r.email || '—',
      headerClassName: 'w-64',
    },
    {
      id: 'phone',
      header: t('members.phone'),
      cell: (r) => r.phone || '—',
      headerClassName: 'w-40',
    },
    {
      id: 'role',
      header: t('members.role'),
      cell: (r) => r.roleName,
      headerClassName: 'w-32',
    },
    {
      id: 'status',
      header: t('members.status'),
      cell: (r) => statusLabel(t, r.status),
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
          <h1 className="text-2xl font-semibold">{t('nav.members')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('members.subtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('members.newMember')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={members ?? []}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('members.noMembers')}
      />

      <MemberDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
    </div>
  )
}

function statusLabel(t: (k: string) => string, status: string): string {
  if (status === 'active') return t('members.statusActive')
  if (status === 'inactive') return t('members.statusInactive')
  if (status === 'suspended') return t('members.statusSuspended')
  return status || '—'
}
