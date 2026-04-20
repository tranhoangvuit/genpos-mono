import { ConnectError } from '@connectrpc/connect'
import { Pencil, Store, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import {
  AllTab,
  Avatar,
  Check,
  DeleteBanner,
  IconBtn,
  LP_FG,
  LP_MUTED_FG,
  ListHeader,
  ListPageShell,
  ListPagination,
  ListSection,
  PrimaryBtn,
  Td,
  Th,
} from '@/shared/ui/list-page'

import { StoreDialog } from './StoreDialog'
import { useDeleteStore, useStores } from './hooks'
import type { StoreRow } from './types'

const PAGE_SIZE = 15

function statusLabel(t: (k: string) => string, status: string): string {
  if (status === 'active') return t('stores.statusActive')
  if (status === 'inactive') return t('stores.statusInactive')
  if (status === 'closed') return t('stores.statusClosed')
  return status || '—'
}

export function StoresPage() {
  const { t } = useTranslation()
  const { data: stores, isLoading } = useStores()
  const deleteMut = useDeleteStore()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<StoreRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<StoreRow | null>(null)
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const list = stores ?? []

  const pageStart = page * PAGE_SIZE
  const pageRows = list.slice(pageStart, pageStart + PAGE_SIZE)
  const pageEnd = pageStart + pageRows.length
  const allSelectedOnPage =
    pageRows.length > 0 && pageRows.every((r) => selected.has(r.id))

  function toggle(id: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }
  function toggleAll() {
    setSelected((prev) => {
      const next = new Set(prev)
      if (allSelectedOnPage) pageRows.forEach((r) => next.delete(r.id))
      else pageRows.forEach((r) => next.add(r.id))
      return next
    })
  }

  const confirmDelete = async () => {
    if (!pendingDelete) return
    try {
      await deleteMut.mutateAsync(pendingDelete.id)
      setPendingDelete(null)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
      setPendingDelete(null)
    }
  }

  return (
    <ListPageShell>
      <ListHeader
        icon={<Store className="h-[18px] w-[18px]" strokeWidth={2} />}
        title={t('nav.stores')}
        count={list.length}
        actions={
          <PrimaryBtn
            onClick={() => {
              setEditing(null)
              setDialogOpen(true)
            }}
          >
            {t('stores.newStore')}
          </PrimaryBtn>
        }
      />

      <DeleteBanner message={deleteError} />

      <ListSection>
        <AllTab count={list.length} />

        <div className="w-full overflow-x-auto">
          <table className="w-full border-collapse" style={{ minWidth: 1000 }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 48 }} />
              <col />
              <col />
              <col style={{ width: 160 }} />
              <col style={{ width: 220 }} />
              <col style={{ width: 110 }} />
              <col style={{ width: 88 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th />
                <Th>{t('stores.name')}</Th>
                <Th>{t('stores.address')}</Th>
                <Th>{t('stores.phone')}</Th>
                <Th>{t('stores.email')}</Th>
                <Th>{t('stores.status')}</Th>
                <Th />
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-[13px]" style={{ color: LP_MUTED_FG }}>
                    {t('common.loading')}
                  </td>
                </tr>
              )}
              {!isLoading && pageRows.length === 0 && (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-[13px]" style={{ color: LP_MUTED_FG }}>
                    {t('stores.noStores')}
                  </td>
                </tr>
              )}
              {pageRows.map((r, i) => (
                <tr key={r.id} className="group transition hover:bg-[hsl(210_40%_96%_/_0.3)]">
                  <Td>
                    <Check checked={selected.has(r.id)} onClick={() => toggle(r.id)} />
                  </Td>
                  <Td>
                    <Avatar name={r.name} index={pageStart + i} />
                  </Td>
                  <Td>
                    <button
                      type="button"
                      onClick={() => {
                        setEditing(r)
                        setDialogOpen(true)
                      }}
                      className="font-medium hover:underline"
                      style={{ color: LP_FG, textDecorationColor: LP_MUTED_FG }}
                    >
                      {r.name}
                    </button>
                  </Td>
                  <Td>
                    <span style={{ color: r.address ? LP_FG : LP_MUTED_FG }}>
                      {r.address || '—'}
                    </span>
                  </Td>
                  <Td>
                    <span className="tabular-nums" style={{ color: r.phone ? LP_FG : LP_MUTED_FG }}>
                      {r.phone || '—'}
                    </span>
                  </Td>
                  <Td>
                    <span style={{ color: r.email ? LP_FG : LP_MUTED_FG }}>{r.email || '—'}</span>
                  </Td>
                  <Td>
                    <span style={{ color: LP_MUTED_FG }}>{statusLabel(t, r.status)}</span>
                  </Td>
                  <Td align="right">
                    <div className="flex items-center justify-end gap-1 opacity-0 transition group-hover:opacity-100">
                      <IconBtn
                        onClick={() => {
                          setEditing(r)
                          setDialogOpen(true)
                        }}
                        label={t('common.edit')}
                      >
                        <Pencil className="h-3.5 w-3.5" strokeWidth={2} />
                      </IconBtn>
                      <IconBtn
                        onClick={() => {
                          setDeleteError(null)
                          setPendingDelete(r)
                        }}
                        label={t('common.delete')}
                      >
                        <Trash2 className="h-3.5 w-3.5" strokeWidth={2} />
                      </IconBtn>
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <ListPagination
          pageStart={pageStart}
          pageEnd={pageEnd}
          total={list.length}
          page={page}
          onPrev={() => setPage((p) => Math.max(0, p - 1))}
          onNext={() => setPage((p) => p + 1)}
        />
      </ListSection>

      <StoreDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />

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
          <p className="text-sm" style={{ color: LP_MUTED_FG }}>
            {pendingDelete ? t('stores.confirmDelete', { name: pendingDelete.name }) : ''}
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
    </ListPageShell>
  )
}
