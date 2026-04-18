import { ConnectError } from '@connectrpc/connect'
import { useNavigate } from '@tanstack/react-router'
import { Ban, CheckCircle2, Save, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

import {
  useCancelStockTake,
  useDeleteStockTake,
  useFinalizeStockTake,
  useSaveStockTakeProgress,
  useStockTakeItems,
  useStockTakeRow,
  useStores,
  useVariantPicker,
} from './hooks'

const STATUS_STYLE: Record<string, string> = {
  in_progress: 'bg-blue-500/15 text-blue-600 dark:text-blue-400',
  completed: 'bg-[color:var(--color-success)]/15 text-[color:var(--color-success)]',
  cancelled: 'bg-[color:var(--color-destructive)]/15 text-[color:var(--color-destructive)]',
}

type Props = { takeId: string }

export function StockTakeDetail({ takeId }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const { data: takeRows } = useStockTakeRow(takeId)
  const { data: itemRows } = useStockTakeItems(takeId)
  const { data: stores } = useStores()
  const { data: variants } = useVariantPicker()

  const save = useSaveStockTakeProgress()
  const finalize = useFinalizeStockTake()
  const cancel = useCancelStockTake()
  const del = useDeleteStockTake()

  const take = takeRows?.[0]
  const variantById = useMemo(
    () => new Map((variants ?? []).map((v) => [v.id, v])),
    [variants],
  )
  const storeName = stores?.find((s) => s.id === take?.store_id)?.name ?? '—'

  // Local draft state keyed by item.id → counted_qty string.
  const [counts, setCounts] = useState<Record<string, string>>({})
  const [notes, setNotes] = useState('')
  const [filter, setFilter] = useState('')
  const [actionError, setActionError] = useState<string | null>(null)

  useEffect(() => {
    const next: Record<string, string> = {}
    for (const it of itemRows ?? []) {
      next[it.id] = it.counted_qty
    }
    setCounts(next)
  }, [itemRows])

  useEffect(() => {
    if (take && take.notes !== null) setNotes(take.notes)
  }, [take])

  if (!take) {
    return (
      <div className="text-sm text-[color:var(--color-muted-foreground)]">
        {t('common.loading')}
      </div>
    )
  }

  const readOnly = take.status !== 'in_progress'
  const needle = filter.trim().toLowerCase()
  const visibleItems = (itemRows ?? []).filter((it) => {
    if (needle === '') return true
    const v = variantById.get(it.variant_id)
    const label = v ? `${v.product_name} ${v.variant_name} ${v.sku ?? ''}` : ''
    return label.toLowerCase().includes(needle)
  })

  const totals = (itemRows ?? []).reduce(
    (acc, it) => {
      const expected = Number(it.expected_qty)
      const counted = Number(counts[it.id] ?? it.counted_qty)
      if (Number.isNaN(expected) || Number.isNaN(counted)) return acc
      const variance = counted - expected
      return {
        expected: acc.expected + expected,
        counted: acc.counted + counted,
        varianceLines: acc.varianceLines + (variance !== 0 ? 1 : 0),
      }
    },
    { expected: 0, counted: 0, varianceLines: 0 },
  )

  const onSave = async () => {
    setActionError(null)
    const lines = Object.entries(counts).map(([itemId, countedQty]) => ({
      itemId,
      countedQty: countedQty || '0',
    }))
    try {
      await save.mutateAsync({ id: take.id, notes, lines })
    } catch (err) {
      setActionError(ConnectError.from(err).rawMessage)
    }
  }

  const onFinalize = async () => {
    if (!confirm(t('inventory.confirmFinalize'))) return
    setActionError(null)
    try {
      // Save first so the latest counts are captured.
      const lines = Object.entries(counts).map(([itemId, countedQty]) => ({
        itemId,
        countedQty: countedQty || '0',
      }))
      await save.mutateAsync({ id: take.id, notes, lines })
      await finalize.mutateAsync(take.id)
    } catch (err) {
      setActionError(ConnectError.from(err).rawMessage)
    }
  }

  const onCancel = async () => {
    if (!confirm(t('inventory.confirmCancelTake'))) return
    setActionError(null)
    try {
      await cancel.mutateAsync(take.id)
    } catch (err) {
      setActionError(ConnectError.from(err).rawMessage)
    }
  }

  const onDelete = async () => {
    if (!confirm(t('inventory.confirmDeleteTake'))) return
    setActionError(null)
    try {
      await del.mutateAsync(take.id)
      void navigate({ to: '/inventory/stock-takes' })
    } catch (err) {
      setActionError(ConnectError.from(err).rawMessage)
    }
  }

  const busy = save.isPending || finalize.isPending || cancel.isPending || del.isPending

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
            <h1 className="text-2xl font-semibold">
              {t('inventory.stockTakeTitle', { date: formatDate(take.created_at) })}
            </h1>
            <span
              className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
                STATUS_STYLE[take.status] ?? STATUS_STYLE.in_progress
              }`}
            >
              {t(`inventory.takeStatus_${take.status}`, take.status)}
            </span>
          </div>
          <p className="mt-1 text-sm text-[color:var(--color-muted-foreground)]">
            {storeName} · {t('inventory.itemsCount', { n: itemRows?.length ?? 0 })}
          </p>
        </div>
        <div className="flex gap-2">
          {!readOnly && (
            <>
              <Button variant="outline" onClick={onSave} disabled={busy}>
                <Save className="mr-2 h-4 w-4" />
                {t('inventory.saveProgress')}
              </Button>
              <Button onClick={onFinalize} disabled={busy}>
                <CheckCircle2 className="mr-2 h-4 w-4" />
                {t('inventory.finalize')}
              </Button>
              <Button variant="outline" onClick={onCancel} disabled={busy}>
                <Ban className="mr-2 h-4 w-4" />
                {t('inventory.cancel')}
              </Button>
            </>
          )}
          {take.status === 'cancelled' && (
            <Button variant="outline" onClick={onDelete} disabled={busy}>
              <Trash2 className="mr-2 h-4 w-4" />
              {t('common.delete')}
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
        <StatCard label={t('inventory.items')} value={String(itemRows?.length ?? 0)} />
        <StatCard label={t('inventory.expectedTotal')} value={totals.expected.toLocaleString()} />
        <StatCard label={t('inventory.countedTotal')} value={totals.counted.toLocaleString()} />
        <StatCard label={t('inventory.varianceLines')} value={String(totals.varianceLines)} />
      </div>

      <div className="max-w-sm">
        <Input
          placeholder={t('inventory.searchVariant')}
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </div>

      <div className="rounded-lg border border-[color:var(--color-border)]">
        <table className="w-full text-sm">
          <thead className="bg-[color:var(--color-muted)]/40 text-xs text-[color:var(--color-muted-foreground)]">
            <tr>
              <th className="px-3 py-2 text-left">{t('inventory.variant')}</th>
              <th className="px-3 py-2 text-left">{t('inventory.sku')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.expected')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.counted')}</th>
              <th className="px-3 py-2 text-right">{t('inventory.variance')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[color:var(--color-border)]">
            {visibleItems.map((it) => {
              const v = variantById.get(it.variant_id)
              const expected = Number(it.expected_qty)
              const countedStr = counts[it.id] ?? it.counted_qty
              const counted = Number(countedStr)
              const variance =
                Number.isNaN(expected) || Number.isNaN(counted) ? 0 : counted - expected
              return (
                <tr key={it.id}>
                  <td className="px-3 py-2">
                    {v ? `${v.product_name} / ${v.variant_name}` : it.variant_id}
                  </td>
                  <td className="px-3 py-2 text-[color:var(--color-muted-foreground)]">
                    {v?.sku ?? ''}
                  </td>
                  <td className="px-3 py-2 text-right">{it.expected_qty}</td>
                  <td className="px-3 py-2">
                    <Input
                      type="number"
                      step="0.0001"
                      className="text-right"
                      value={countedStr}
                      disabled={readOnly}
                      onChange={(e) =>
                        setCounts((c) => ({ ...c, [it.id]: e.target.value }))
                      }
                    />
                  </td>
                  <td
                    className={`px-3 py-2 text-right ${
                      variance === 0
                        ? ''
                        : variance > 0
                          ? 'text-[color:var(--color-success)]'
                          : 'text-[color:var(--color-destructive)]'
                    }`}
                  >
                    {variance > 0 ? `+${variance}` : variance}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      <div className="space-y-2 max-w-2xl">
        <Label htmlFor="notes">{t('inventory.notes')}</Label>
        <Textarea
          id="notes"
          rows={3}
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          disabled={readOnly}
        />
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4">
      <div className="text-xs text-[color:var(--color-muted-foreground)]">{label}</div>
      <div className="mt-1 text-xl font-semibold">{value}</div>
    </div>
  )
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}
