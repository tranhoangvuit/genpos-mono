import { ConnectError } from '@connectrpc/connect'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'

import { useReceivePurchaseOrder } from './hooks'
import type { PurchaseOrderItemRow, VariantPickerRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  poId: string
  items: PurchaseOrderItemRow[]
  variants: VariantPickerRow[]
}

export function ReceiveDialog({ open, onOpenChange, poId, items, variants }: Props) {
  const { t } = useTranslation()
  const receive = useReceivePurchaseOrder()

  const [deltas, setDeltas] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    // Pre-fill each row with remaining qty.
    const next: Record<string, string> = {}
    for (const it of items) {
      const remaining = Number(it.quantity_ordered) - Number(it.quantity_received)
      next[it.id] = remaining > 0 ? String(remaining) : '0'
    }
    setDeltas(next)
    setError(null)
    receive.reset()
  }, [open, items, receive])

  const variantById = new Map(variants.map((v) => [v.id, v]))

  const onReceive = async () => {
    setError(null)
    const lines = Object.entries(deltas)
      .filter(([, v]) => Number(v) > 0)
      .map(([itemId, qty]) => ({ itemId, quantityToReceive: qty }))
    if (lines.length === 0) {
      setError(t('inventory.validation.nothingToReceive'))
      return
    }
    try {
      await receive.mutateAsync({ id: poId, lines })
      onOpenChange(false)
    } catch (err) {
      setError(ConnectError.from(err).rawMessage)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('inventory.receiveStock')}</DialogTitle>
          <DialogDescription>{t('inventory.receiveSubtitle')}</DialogDescription>
        </DialogHeader>

        {error && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {error}
          </div>
        )}

        <div className="max-h-96 overflow-auto rounded-lg border border-[color:var(--color-border)]">
          <table className="w-full text-sm">
            <thead className="bg-[color:var(--color-muted)]/40 text-xs text-[color:var(--color-muted-foreground)]">
              <tr>
                <th className="px-3 py-2 text-left">{t('inventory.variant')}</th>
                <th className="px-3 py-2 text-right">{t('inventory.ordered')}</th>
                <th className="px-3 py-2 text-right">{t('inventory.received')}</th>
                <th className="px-3 py-2 text-right">{t('inventory.receiving')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[color:var(--color-border)]">
              {items.map((it) => {
                const v = variantById.get(it.variant_id)
                const remaining =
                  Number(it.quantity_ordered) - Number(it.quantity_received)
                return (
                  <tr key={it.id}>
                    <td className="px-3 py-2">
                      {v ? `${v.product_name} / ${v.variant_name}` : it.variant_id}
                    </td>
                    <td className="px-3 py-2 text-right">{it.quantity_ordered}</td>
                    <td className="px-3 py-2 text-right">{it.quantity_received}</td>
                    <td className="px-3 py-2">
                      <Input
                        type="number"
                        step="0.0001"
                        min="0"
                        max={String(remaining)}
                        className="text-right"
                        value={deltas[it.id] ?? ''}
                        onChange={(e) =>
                          setDeltas((d) => ({ ...d, [it.id]: e.target.value }))
                        }
                      />
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={receive.isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button type="button" onClick={onReceive} disabled={receive.isPending}>
            {receive.isPending ? t('common.saving') : t('inventory.confirmReceive')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
