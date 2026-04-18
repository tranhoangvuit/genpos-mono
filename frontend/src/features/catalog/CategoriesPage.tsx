import { ConnectError } from '@connectrpc/connect'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'

import { CategoryDialog } from './CategoryDialog'
import { useCategories, useDeleteCategory } from './hooks'
import type { CategoryRow } from './types'

export function CategoriesPage() {
  const { t } = useTranslation()
  const { data: categories, isLoading } = useCategories()
  const deleteMut = useDeleteCategory()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<CategoryRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<CategoryRow | null>(null)

  const list = categories ?? []
  const parents = list

  const onEdit = (row: CategoryRow) => {
    setEditing(row)
    setDialogOpen(true)
  }

  const onDelete = (row: CategoryRow) => {
    setDeleteError(null)
    setPendingDelete(row)
  }

  const confirmDelete = async () => {
    if (!pendingDelete) return
    const row = pendingDelete
    try {
      await deleteMut.mutateAsync(row.id)
      setPendingDelete(null)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
      setPendingDelete(null)
    }
  }

  const columns: DataTableColumn<CategoryRow>[] = [
    {
      id: 'name',
      header: t('catalog.name'),
      cell: (r) => (
        <span className="flex items-center gap-2">
          {r.color && (
            <span
              className="h-3 w-3 rounded-full border border-[color:var(--color-border)]"
              style={{ backgroundColor: r.color }}
              aria-hidden
            />
          )}
          <span className="font-medium">{r.name}</span>
        </span>
      ),
    },
    {
      id: 'parent',
      header: t('catalog.parentCategory'),
      cell: (r) => list.find((p) => p.id === r.parentId)?.name ?? '—',
    },
    {
      id: 'sortOrder',
      header: t('catalog.sortOrder'),
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
          <h1 className="text-2xl font-semibold">{t('nav.categories')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('catalog.categoriesSubtitle')}
          </p>
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('catalog.newCategory')}
        </Button>
      </div>

      {deleteError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {deleteError}
        </div>
      )}

      <DataTable
        columns={columns}
        data={list}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('catalog.noCategories')}
      />

      <CategoryDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        existing={editing}
        parents={parents}
      />

      <Dialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) setPendingDelete(null)
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('common.delete')}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {pendingDelete
              ? t('catalog.confirmDeleteCategory', { name: pendingDelete.name })
              : ''}
          </p>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setPendingDelete(null)}
              disabled={deleteMut.isPending}
            >
              {t('common.cancel')}
            </Button>
            <Button
              type="button"
              variant="destructive"
              onClick={confirmDelete}
              disabled={deleteMut.isPending}
            >
              {deleteMut.isPending ? t('common.saving') : t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
