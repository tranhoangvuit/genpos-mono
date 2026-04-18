import { ConnectError } from '@connectrpc/connect'
import { useNavigate } from '@tanstack/react-router'
import { Ban, PackageCheck, Send, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'

import { PurchaseOrderForm } from './PurchaseOrderForm'
import { ReceiveDialog } from './ReceiveDialog'
import {
  useCancelPurchaseOrder,
  useDeletePurchaseOrder,
  usePurchaseOrder,
  useSubmitPurchaseOrder,
  useVariantPicker,
} from './hooks'

const STATUS_STYLE: Record<string, string> = {
  draft: 'bg-[color:var(--color-muted)] text-[color:var(--color-muted-foreground)]',
  submitted: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  partial: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  received: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  cancelled: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
}

type Props = { poId: string }

export function PurchaseOrderDetail({ poId }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: po } = usePurchaseOrder(poId)
  const itemRows = po?.items ?? []
  const { data: variants } = useVariantPicker()

  const submitMut = useSubmitPurchaseOrder()
  const cancelMut = useCancelPurchaseOrder()
  const deleteMut = useDeletePurchaseOrder()

  const [actionError, setActionError] = useState<string | null>(null)
  const [receiveOpen, setReceiveOpen] = useState(false)

  if (!po) {
    return <div className="text-sm text-[color:var(--color-muted-foreground)]">{t('common.loading')}</div>
  }

  // Draft POs are rendered via the form (read-only guard is internal).
  if (po.status === 'draft') {
    return (
      <div className="space-y-4">
        <DraftActions
          poId={po.id}
          onSubmit={async () => {
            setActionError(null)
            try {
              await submitMut.mutateAsync(po.id)
            } catch (err) {
              setActionError(ConnectError.from(err).rawMessage)
            }
          }}
          onDelete={async () => {
            if (!confirm(t('inventory.confirmDelete', { number: po.poNumber }))) return
            setActionError(null)
            try {
              await deleteMut.mutateAsync(po.id)
              void navigate({ to: '/inventory/purchase-orders' })
            } catch (err) {
              setActionError(ConnectError.from(err).rawMessage)
            }
          }}
          submitting={submitMut.isPending}
          deleting={deleteMut.isPending}
          error={actionError}
        />
        <PurchaseOrderForm poId={po.id} />
      </div>
    )
  }

  // Non-draft: read-only view with action toolbar
  const variantById = new Map((variants ?? []).map((v) => [v.id, v]))
  const total = itemRows.reduce((sum, it) => {
    const q = Number(it.quantityOrdered)
    const c = Number(it.costPrice)
    return sum + (Number.isNaN(q) || Number.isNaN(c) ? 0 : q * c)
  }, 0)
  const canReceive = po.status === 'submitted' || po.status === 'partial'
  const canCancel = po.status !== 'received' && po.status !== 'cancelled'
  const canDelete = po.status === 'cancelled'

  return (
    <div className="space-y-4">
      {actionError && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {actionError}
        </div>
      )}

      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold">{po.poNumber}</h1>
            <span
              className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
                STATUS_STYLE[po.status] ?? STATUS_STYLE.draft
              }`}
            >
              {t(`inventory.status_${po.status}`, po.status)}
            </span>
          </div>
          <p className="mt-1 text-sm text-[color:var(--color-muted-foreground)]">
            {po.supplierName || '—'}
          </p>
        </div>
        <div className="flex gap-2">
          {canReceive && (
            <Button onClick={() => setReceiveOpen(true)}>
              <PackageCheck className="mr-2 h-4 w-4" />
              {t('inventory.receive')}
            </Button>
          )}
          {canCancel && (
            <Button
              variant="outline"
              onClick={async () => {
                if (!confirm(t('inventory.confirmCancel', { number: po.poNumber }))) return
                setActionError(null)
                try {
                  await cancelMut.mutateAsync(po.id)
                } catch (err) {
                  setActionError(ConnectError.from(err).rawMessage)
                }
              }}
              disabled={cancelMut.isPending}
            >
              <Ban className="mr-2 h-4 w-4" />
              {t('inventory.cancel')}
            </Button>
          )}
          {canDelete && (
            <Button
              variant="outline"
              onClick={async () => {
                if (!confirm(t('inventory.confirmDelete', { number: po.poNumber }))) return
                setActionError(null)
                try {
                  await deleteMut.mutateAsync(po.id)
                  void navigate({ to: '/inventory/purchase-orders' })
                } catch (err) {
                  setActionError(ConnectError.from(err).rawMessage)
                }
              }}
              disabled={deleteMut.isPending}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {t('common.delete')}
            </Button>
          )}
        </div>
      </div>

      <div className="rounded-lg border border-[color:var(--color-border)]">
        <table className="w-full text-sm">
          <thead className="bg-[color:var(--color-muted)]/40 text-xs text-[color:var(--color-muted-foreground)]">
            <tr>
              <th className="px-3 py-2 text-left">{t('inventory.variant')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.ordered')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.received')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.unitCost')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.lineTotal')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[color:var(--color-border)]">
            {itemRows.map((it) => {
              const v = variantById.get(it.variantId)
              const ordered = Number(it.quantityOrdered)
              const received = Number(it.quantityReceived)
              const cost = Number(it.costPrice)
              const line = Number.isNaN(ordered) || Number.isNaN(cost) ? 0 : ordered * cost
              return (
                <tr key={it.id}>
                  <td className="px-3 py-2">
                    {v ? `${v.productName} / ${v.variantName}` : it.variantId}
                  </td>
                  <td className="px-3 py-2 text-right">{ordered}</td>
                  <td className="px-3 py-2 text-right">
                    <span className={received < ordered ? 'text-amber-600 dark:text-amber-400' : ''}>
                      {received}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-right">{cost.toLocaleString()}</td>
                  <td className="px-3 py-2 text-right">{line.toLocaleString()}</td>
                </tr>
              )
            })}
          </tbody>
          <tfoot>
            <tr className="border-t border-[color:var(--color-border)] font-medium">
              <td className="px-3 py-2 text-right" colSpan={4}>
                {t('inventory.total')}
              </td>
              <td className="px-3 py-2 text-right">{total.toLocaleString()}</td>
            </tr>
          </tfoot>
        </table>
      </div>

      {po.notes && (
        <div>
          <div className="text-xs font-medium text-[color:var(--color-muted-foreground)]">
            {t('inventory.notes')}
          </div>
          <p className="mt-1 whitespace-pre-wrap text-sm">{po.notes}</p>
        </div>
      )}

      <ReceiveDialog
        open={receiveOpen}
        onOpenChange={setReceiveOpen}
        poId={po.id}
        items={itemRows}
        variants={variants ?? []}
      />
    </div>
  )
}

type DraftActionsProps = {
  poId: string
  onSubmit: () => void
  onDelete: () => void
  submitting: boolean
  deleting: boolean
  error: string | null
}

function DraftActions({ onSubmit, onDelete, submitting, deleting, error }: DraftActionsProps) {
  const { t } = useTranslation()
  return (
    <>
      {error && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {error}
        </div>
      )}
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onDelete} disabled={deleting}>
          <Trash2 className="mr-2 h-4 w-4" />
          {t('common.delete')}
        </Button>
        <Button onClick={onSubmit} disabled={submitting}>
          <Send className="mr-2 h-4 w-4" />
          {t('inventory.submit')}
        </Button>
      </div>
    </>
  )
}
