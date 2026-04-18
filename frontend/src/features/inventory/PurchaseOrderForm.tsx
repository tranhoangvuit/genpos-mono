import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { ConnectError } from '@connectrpc/connect'
import { useNavigate } from '@tanstack/react-router'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { Plus, Trash2 } from 'lucide-react'
import { useEffect, useMemo } from 'react'
import { useFieldArray, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

import {
  useCreatePurchaseOrder,
  usePurchaseOrderItems,
  usePurchaseOrderRow,
  useStores,
  useSuppliers,
  useUpdatePurchaseOrder,
  useVariantPicker,
} from './hooks'
import {
  purchaseOrderSchema,
  type PurchaseOrderFormData,
} from './schemas'

type Props = {
  poId?: string // edit mode if set
}

export function PurchaseOrderForm({ poId }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const create = useCreatePurchaseOrder()
  const update = useUpdatePurchaseOrder()

  const { data: stores } = useStores()
  const { data: suppliers } = useSuppliers()
  const { data: variants } = useVariantPicker()
  const { data: poRows } = usePurchaseOrderRow(poId)
  const { data: poItemRows } = usePurchaseOrderItems(poId)

  const existing = poRows?.[0]

  const form = useForm<PurchaseOrderFormData>({
    resolver: standardSchemaResolver(purchaseOrderSchema(t)),
    defaultValues: {
      storeId: '',
      supplierName: '',
      notes: '',
      expectedAt: '',
      items: [],
    },
  })
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'items',
  })

  // Default to first store for new POs.
  useEffect(() => {
    if (!existing && !form.getValues('storeId') && stores && stores.length > 0) {
      form.setValue('storeId', stores[0].id)
    }
  }, [existing, stores, form])

  // Populate edit form when underlying rows arrive.
  useEffect(() => {
    if (!existing) return
    const items = (poItemRows ?? []).map((it) => {
      const v = variants?.find((x) => x.id === it.variant_id)
      return {
        variantId: it.variant_id,
        variantLabel: v ? `${v.product_name} / ${v.variant_name}` : it.variant_id,
        quantityOrdered: it.quantity_ordered,
        costPrice: it.cost_price,
      }
    })
    form.reset({
      storeId: existing.store_id,
      supplierName: existing.supplier_name ?? '',
      notes: existing.notes ?? '',
      expectedAt: existing.expected_at ? existing.expected_at.slice(0, 10) : '',
      items,
    })
  }, [existing, poItemRows, variants, form])

  const variantOptions = useMemo(
    () =>
      (variants ?? []).map((v) => ({
        id: v.id,
        label: `${v.product_name} / ${v.variant_name}${v.sku ? ` (${v.sku})` : ''}`,
        costPrice: v.cost_price,
      })),
    [variants],
  )

  const submitting = create.isPending || update.isPending
  const serverError = create.error ?? update.error
  const errorMessage = serverError ? ConnectError.from(serverError).rawMessage : null

  const readOnly = existing ? existing.status !== 'draft' : false

  const total = (form.watch('items') ?? []).reduce((sum, it) => {
    const q = Number(it.quantityOrdered)
    const c = Number(it.costPrice)
    if (Number.isNaN(q) || Number.isNaN(c)) return sum
    return sum + q * c
  }, 0)

  const onSubmit = form.handleSubmit(async (values) => {
    const expectedAt = values.expectedAt
      ? timestampFromDate(new Date(values.expectedAt + 'T00:00:00'))
      : undefined
    const payload = {
      purchaseOrder: {
        storeId: values.storeId,
        supplierName: values.supplierName,
        notes: values.notes,
        expectedAt,
        items: values.items.map((it) => ({
          variantId: it.variantId,
          quantityOrdered: it.quantityOrdered,
          costPrice: it.costPrice || '0',
        })),
      },
    }
    let res
    if (existing) {
      res = await update.mutateAsync({ id: existing.id, ...payload })
    } else {
      res = await create.mutateAsync(payload)
    }
    const id = res.purchaseOrder?.id
    if (id) {
      void navigate({ to: '/inventory/purchase-orders/$id', params: { id } })
    } else {
      void navigate({ to: '/inventory/purchase-orders' })
    }
  })

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-semibold">
          {existing
            ? t('inventory.editPurchaseOrder', { number: existing.po_number })
            : t('inventory.newPurchaseOrder')}
        </h1>
        <p className="text-sm text-[color:var(--color-muted-foreground)]">
          {existing && readOnly
            ? t('inventory.readOnlyHint')
            : t('inventory.purchaseOrderFormSubtitle')}
        </p>
      </div>

      {errorMessage && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {errorMessage}
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-6" noValidate>
        <fieldset disabled={readOnly} className="space-y-6">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="storeId">{t('inventory.store')}</Label>
              <select
                id="storeId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('storeId')}
              >
                <option value="">—</option>
                {(stores ?? []).map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
              {form.formState.errors.storeId && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.storeId.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="supplierName">{t('inventory.supplier')}</Label>
              <Input
                id="supplierName"
                list="supplier-options"
                placeholder={t('inventory.supplierPlaceholder')}
                {...form.register('supplierName')}
              />
              <datalist id="supplier-options">
                {(suppliers ?? []).map((s) => (
                  <option key={s.id} value={s.name} />
                ))}
              </datalist>
            </div>

            <div className="space-y-2">
              <Label htmlFor="expectedAt">{t('inventory.expected')}</Label>
              <Input id="expectedAt" type="date" {...form.register('expectedAt')} />
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>{t('inventory.items')}</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  append({
                    variantId: '',
                    variantLabel: '',
                    quantityOrdered: '1',
                    costPrice: '0',
                  })
                }
              >
                <Plus className="mr-2 h-4 w-4" />
                {t('inventory.addItem')}
              </Button>
            </div>
            {form.formState.errors.items?.root && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.items.root.message}
              </p>
            )}
            {fields.length === 0 ? (
              <p className="text-sm text-[color:var(--color-muted-foreground)]">
                {t('inventory.noItemsHint')}
              </p>
            ) : (
              <div className="rounded-lg border border-[color:var(--color-border)]">
                <table className="w-full text-sm">
                  <thead className="bg-[color:var(--color-muted)]/40 text-xs text-[color:var(--color-muted-foreground)]">
                    <tr>
                      <th className="px-3 py-2 text-left">{t('inventory.variant')}</th>
                      <th className="px-3 py-2 text-right">{t('inventory.qty')}</th>
                      <th className="px-3 py-2 text-right">{t('inventory.unitCost')}</th>
                      <th className="px-3 py-2 text-right">{t('inventory.lineTotal')}</th>
                      <th className="w-12"></th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[color:var(--color-border)]">
                    {fields.map((f, i) => {
                      const qty = Number(form.watch(`items.${i}.quantityOrdered`) || '0')
                      const unit = Number(form.watch(`items.${i}.costPrice`) || '0')
                      const line = Number.isNaN(qty) || Number.isNaN(unit) ? 0 : qty * unit
                      return (
                        <tr key={f.id}>
                          <td className="px-3 py-2">
                            <select
                              className="h-9 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-2 text-sm"
                              value={form.watch(`items.${i}.variantId`)}
                              onChange={(e) => {
                                const v = variantOptions.find((x) => x.id === e.target.value)
                                form.setValue(`items.${i}.variantId`, e.target.value)
                                form.setValue(`items.${i}.variantLabel`, v?.label ?? '')
                                if (v && !form.getValues(`items.${i}.costPrice`)) {
                                  form.setValue(`items.${i}.costPrice`, v.costPrice)
                                }
                              }}
                            >
                              <option value="">{t('inventory.selectVariant')}</option>
                              {variantOptions.map((v) => (
                                <option key={v.id} value={v.id}>
                                  {v.label}
                                </option>
                              ))}
                            </select>
                            {form.formState.errors.items?.[i]?.variantId && (
                              <p className="mt-1 text-xs text-[color:var(--color-destructive)]">
                                {form.formState.errors.items[i]?.variantId?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-3 py-2">
                            <Input
                              type="number"
                              step="0.0001"
                              className="text-right"
                              {...form.register(`items.${i}.quantityOrdered`)}
                            />
                          </td>
                          <td className="px-3 py-2">
                            <Input
                              type="number"
                              step="0.0001"
                              className="text-right"
                              {...form.register(`items.${i}.costPrice`)}
                            />
                          </td>
                          <td className="px-3 py-2 text-right">{line.toLocaleString()}</td>
                          <td className="px-3 py-2 text-right">
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => remove(i)}
                              aria-label={t('common.remove')}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                  <tfoot>
                    <tr className="border-t border-[color:var(--color-border)] font-medium">
                      <td className="px-3 py-2 text-right" colSpan={3}>
                        {t('inventory.total')}
                      </td>
                      <td className="px-3 py-2 text-right">{total.toLocaleString()}</td>
                      <td></td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">{t('inventory.notes')}</Label>
            <Textarea id="notes" rows={3} {...form.register('notes')} />
          </div>
        </fieldset>

        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate({ to: '/inventory/purchase-orders' })}
            disabled={submitting}
          >
            {t('common.cancel')}
          </Button>
          {!readOnly && (
            <Button type="submit" disabled={submitting}>
              {submitting
                ? t('common.saving')
                : existing
                  ? t('common.save')
                  : t('common.create')}
            </Button>
          )}
        </div>
      </form>
    </div>
  )
}
