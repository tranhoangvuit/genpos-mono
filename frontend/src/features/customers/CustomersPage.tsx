import { ConnectError } from '@connectrpc/connect'
import { Pencil, Trash2, Upload, Users } from 'lucide-react'
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
  MoreBtn,
  PrimaryBtn,
  Td,
  Th,
} from '@/shared/ui/list-page'

import { CustomerDialog } from './CustomerDialog'
import { ImportCustomerDialog } from './ImportCustomerDialog'
import { useCustomers, useDeleteCustomer, useGetCustomer } from './hooks'
import type { CustomerListRow, CustomerRow } from './types'

const PAGE_SIZE = 15

export function CustomersPage() {
  const { t } = useTranslation()
  const { data: customers, isLoading } = useCustomers()
  const deleteMut = useDeleteCustomer()
  const getMut = useGetCustomer()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<CustomerRow | null>(null)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<CustomerListRow | null>(null)
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const list = customers ?? []

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

  const onEdit = async (r: CustomerListRow) => {
    const res = await getMut.mutateAsync(r.id)
    if (res.customer) {
      setEditing(res.customer)
      setDialogOpen(true)
    }
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
        icon={<Users className="h-[18px] w-[18px]" strokeWidth={2} />}
        title={t('nav.customers')}
        count={list.length}
        actions={
          <>
            <MoreBtn icon={<Upload className="h-3.5 w-3.5" />} onClick={() => setImportOpen(true)}>
              {t('customers.importCustomers')}
            </MoreBtn>
            <PrimaryBtn
              onClick={() => {
                setEditing(null)
                setDialogOpen(true)
              }}
            >
              {t('customers.newCustomer')}
            </PrimaryBtn>
          </>
        }
      />

      <DeleteBanner message={deleteError} />

      <ListSection>
        <AllTab count={list.length} />

        <div className="w-full overflow-x-auto">
          <table className="w-full border-collapse" style={{ minWidth: 960 }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 48 }} />
              <col />
              <col style={{ width: 240 }} />
              <col style={{ width: 160 }} />
              <col style={{ width: 200 }} />
              <col style={{ width: 88 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th />
                <Th>{t('customers.name')}</Th>
                <Th>{t('customers.email')}</Th>
                <Th>{t('customers.phone')}</Th>
                <Th>{t('customers.groups')}</Th>
                <Th />
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-[13px]" style={{ color: LP_MUTED_FG }}>
                    {t('common.loading')}
                  </td>
                </tr>
              )}
              {!isLoading && pageRows.length === 0 && (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-[13px]" style={{ color: LP_MUTED_FG }}>
                    {t('customers.noCustomers')}
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
                      onClick={() => onEdit(r)}
                      className="font-medium hover:underline"
                      style={{ color: LP_FG, textDecorationColor: LP_MUTED_FG }}
                    >
                      {r.name}
                    </button>
                  </Td>
                  <Td>
                    <span style={{ color: r.email ? LP_FG : LP_MUTED_FG }}>{r.email || '—'}</span>
                  </Td>
                  <Td>
                    <span className="tabular-nums" style={{ color: r.phone ? LP_FG : LP_MUTED_FG }}>
                      {r.phone || '—'}
                    </span>
                  </Td>
                  <Td>
                    <span style={{ color: r.groupNames ? LP_FG : LP_MUTED_FG }}>
                      {r.groupNames || '—'}
                    </span>
                  </Td>
                  <Td align="right">
                    <div className="flex items-center justify-end gap-1 opacity-0 transition group-hover:opacity-100">
                      <IconBtn onClick={() => onEdit(r)} label={t('common.edit')}>
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

      <CustomerDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />
      <ImportCustomerDialog open={importOpen} onOpenChange={setImportOpen} />

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
            {pendingDelete ? t('customers.confirmDelete', { name: pendingDelete.name }) : ''}
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

