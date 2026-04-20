import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { ConnectError } from '@connectrpc/connect'
import { useNavigate } from '@tanstack/react-router'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import {
  CalendarDays,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Package,
  Plus,
  Search,
  Trash2,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useFieldArray, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'

import {
  useCreatePurchaseOrder,
  usePurchaseOrder,
  useStores,
  useSuppliers,
  useUpdatePurchaseOrder,
  useVariantPicker,
} from './hooks'
import {
  purchaseOrderSchema,
  type PurchaseOrderFormData,
} from './schemas'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const BORDER_STRONG = 'hsl(214 32% 85%)'
const BORDER_HOVER = 'hsl(214 28% 72%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const PRIMARY = 'hsl(221 83% 53%)'
const PRIMARY_INK = 'hsl(224 76% 48%)'
const DESTRUCTIVE = 'hsl(0 84% 60%)'

type Props = {
  poId?: string
}

export function PurchaseOrderForm({ poId }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const create = useCreatePurchaseOrder()
  const update = useUpdatePurchaseOrder()

  const { data: stores } = useStores()
  const { data: suppliers } = useSuppliers()
  const { data: variants } = useVariantPicker()
  const { data: existing } = usePurchaseOrder(poId)

  const [prodSearch, setProdSearch] = useState('')
  const [refNumber, setRefNumber] = useState('')
  const [paymentTerms, setPaymentTerms] = useState('none')
  const [currency, setCurrency] = useState('USD')
  const [carrier, setCarrier] = useState('')
  const [trackingNumber, setTrackingNumber] = useState('')
  const [shippingCost, setShippingCost] = useState('0')

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

  useEffect(() => {
    if (!existing && !form.getValues('storeId') && stores && stores.length > 0) {
      form.setValue('storeId', stores[0].id)
    }
  }, [existing, stores, form])

  useEffect(() => {
    if (!existing) return
    const items = existing.items.map((it) => {
      const v = variants?.find((x) => x.id === it.variantId)
      return {
        variantId: it.variantId,
        variantLabel: v ? `${v.productName} / ${v.variantName}` : it.variantId,
        quantityOrdered: it.quantityOrdered,
        costPrice: it.costPrice,
      }
    })
    const expectedAt = existing.expectedAt
      ? new Date(Number(existing.expectedAt.seconds) * 1000)
          .toISOString()
          .slice(0, 10)
      : ''
    form.reset({
      storeId: existing.storeId,
      supplierName: existing.supplierName ?? '',
      notes: existing.notes ?? '',
      expectedAt,
      items,
    })
  }, [existing, variants, form])

  const variantOptions = useMemo(
    () =>
      (variants ?? []).map((v) => ({
        id: v.id,
        label: `${v.productName} / ${v.variantName}${v.sku ? ` (${v.sku})` : ''}`,
        productName: v.productName,
        variantName: v.variantName,
        sku: v.sku,
        costPrice: v.costPrice,
      })),
    [variants],
  )

  const filteredVariants = useMemo(() => {
    const q = prodSearch.trim().toLowerCase()
    if (!q) return variantOptions.slice(0, 10)
    return variantOptions
      .filter(
        (v) =>
          v.label.toLowerCase().includes(q) || v.sku?.toLowerCase().includes(q),
      )
      .slice(0, 10)
  }, [variantOptions, prodSearch])

  const submitting = create.isPending || update.isPending
  const serverError = create.error ?? update.error
  const errorMessage = serverError ? ConnectError.from(serverError).rawMessage : null

  const readOnly = existing ? existing.status !== 'draft' : false

  const watchedItems = form.watch('items') ?? []
  const subtotal = watchedItems.reduce((sum, it) => {
    const q = Number(it.quantityOrdered)
    const c = Number(it.costPrice)
    if (Number.isNaN(q) || Number.isNaN(c)) return sum
    return sum + q * c
  }, 0)
  const shipping = Number(shippingCost) || 0
  const total = subtotal + shipping
  const itemCount = watchedItems.length
  const variantCount = watchedItems.filter((it) => it.variantId).length

  const addVariant = (id: string) => {
    const v = variantOptions.find((x) => x.id === id)
    if (!v) return
    if (fields.some((f) => f.variantId === v.id)) return
    append({
      variantId: v.id,
      variantLabel: v.label,
      quantityOrdered: '1',
      costPrice: v.costPrice || '0',
    })
    setProdSearch('')
  }

  const backToList = () =>
    navigate({
      to: '/$subdomain/inventory/purchase-orders',
      params: { subdomain },
    })

  const onSubmit = form.handleSubmit(async (values) => {
    const expectedAt = values.expectedAt
      ? timestampFromDate(new Date(values.expectedAt + 'T00:00:00'))
      : undefined
    const notesWithRef = [refNumber && `Ref: ${refNumber}`, values.notes]
      .filter(Boolean)
      .join('\n')
    const payload = {
      purchaseOrder: {
        storeId: values.storeId,
        supplierName: values.supplierName,
        notes: notesWithRef,
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
      void navigate({
        to: '/$subdomain/inventory/purchase-orders/$id',
        params: { subdomain, id },
      })
    } else {
      void backToList()
    }
  })

  return (
    <div className="mx-auto -my-6 p-6" style={{ maxWidth: 1100 }}>
      <div className="mb-[18px] flex items-center gap-1">
        <button
          type="button"
          onClick={() => void backToList()}
          className="inline-grid h-7 w-7 place-items-center rounded-md transition"
          style={{ color: MUTED_FG }}
          aria-label={t('common.back', 'Back')}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = MUTED
            e.currentTarget.style.color = FG
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent'
            e.currentTarget.style.color = MUTED_FG
          }}
        >
          <ChevronLeft className="h-[18px] w-[18px]" />
        </button>
        <ChevronRight className="h-[14px] w-[14px]" style={{ color: MUTED_FG }} />
        <h1
          className="m-0 text-[20px] font-semibold tracking-[-0.005em]"
          style={{ color: FG }}
        >
          {existing
            ? t('inventory.editPurchaseOrder', { number: existing.poNumber })
            : t('inventory.newPurchaseOrder', 'Create purchase order')}
        </h1>
      </div>

      {errorMessage && (
        <div
          className="mb-4 rounded-md px-3 py-2 text-sm"
          style={{
            background: 'hsl(0 84% 60% / 0.08)',
            border: `1px solid ${DESTRUCTIVE}`,
            color: DESTRUCTIVE,
          }}
        >
          {errorMessage}
        </div>
      )}

      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-[14px] pb-20">
        <fieldset disabled={readOnly} className="contents">
          <Card>
            <Grid cols={2}>
              <Field label={t('inventory.supplier', 'Supplier')}>
                <input
                  list="supplier-options"
                  placeholder={t('inventory.supplierPlaceholder', 'Select supplier')}
                  {...form.register('supplierName')}
                  className={inputCls}
                  style={inputStyle}
                />
                <datalist id="supplier-options">
                  {(suppliers ?? []).map((s) => (
                    <option key={s.id} value={s.name} />
                  ))}
                </datalist>
              </Field>
              <Field label={t('inventory.destination', 'Destination')}>
                <NativeSelect
                  value={form.watch('storeId')}
                  onChange={(v) => form.setValue('storeId', v)}
                >
                  <option value="">—</option>
                  {(stores ?? []).map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </NativeSelect>
                {form.formState.errors.storeId && (
                  <p className="mt-1 text-xs" style={{ color: DESTRUCTIVE }}>
                    {form.formState.errors.storeId.message}
                  </p>
                )}
              </Field>
            </Grid>
            <Divider />
            <Grid cols={2}>
              <Field
                label={t('inventory.paymentTerms', 'Payment terms')}
                optional
              >
                <NativeSelect value={paymentTerms} onChange={setPaymentTerms}>
                  <option value="none">None</option>
                  <option value="on_receipt">Due on receipt</option>
                  <option value="net_7">Net 7</option>
                  <option value="net_15">Net 15</option>
                  <option value="net_30">Net 30</option>
                  <option value="net_45">Net 45</option>
                  <option value="net_60">Net 60</option>
                </NativeSelect>
              </Field>
              <Field label={t('inventory.currency', 'Supplier currency')}>
                <NativeSelect value={currency} onChange={setCurrency}>
                  <option value="USD">US Dollar (USD $)</option>
                  <option value="SGD">Singapore Dollar (SGD $)</option>
                  <option value="EUR">Euro (EUR €)</option>
                  <option value="GBP">British Pound (GBP £)</option>
                  <option value="JPY">Japanese Yen (JPY ¥)</option>
                  <option value="AUD">Australian Dollar (AUD $)</option>
                </NativeSelect>
              </Field>
            </Grid>
          </Card>

          <Card>
            <CardTitle>{t('inventory.shipmentDetails', 'Shipment details')}</CardTitle>
            <Grid cols={3}>
              <Field label={t('inventory.estimatedArrival', 'Estimated arrival')}>
                <div className="relative">
                  <input
                    type="date"
                    {...form.register('expectedAt')}
                    className={`${inputCls} pr-9`}
                    style={inputStyle}
                  />
                  <CalendarDays
                    className="pointer-events-none absolute right-3 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
                    style={{ color: MUTED_FG }}
                  />
                </div>
              </Field>
              <Field label={t('inventory.carrier', 'Shipping carrier')}>
                <NativeSelect value={carrier} onChange={setCarrier}>
                  <option value="">Select carrier</option>
                  <option value="dhl">DHL Express</option>
                  <option value="fedex">FedEx</option>
                  <option value="ups">UPS</option>
                  <option value="usps">USPS</option>
                  <option value="singpost">SingPost</option>
                  <option value="other">Other</option>
                </NativeSelect>
              </Field>
              <Field label={t('inventory.trackingNumber', 'Tracking number')}>
                <input
                  type="text"
                  value={trackingNumber}
                  onChange={(e) => setTrackingNumber(e.target.value)}
                  className={inputCls}
                  style={inputStyle}
                />
              </Field>
            </Grid>
          </Card>

          <Card>
            <CardTitle>{t('inventory.addProducts', 'Add products')}</CardTitle>
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Search
                  className="pointer-events-none absolute left-3 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
                  style={{ color: MUTED_FG }}
                />
                <input
                  type="text"
                  list="variant-options"
                  placeholder={t('inventory.searchProducts', 'Search products')}
                  value={prodSearch}
                  onChange={(e) => setProdSearch(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key !== 'Enter') return
                    e.preventDefault()
                    const match = filteredVariants.find(
                      (v) => v.label === prodSearch || v.sku === prodSearch,
                    )
                    if (match) addVariant(match.id)
                  }}
                  className={`${inputCls} pl-9`}
                  style={inputStyle}
                />
                <datalist id="variant-options">
                  {filteredVariants.map((v) => (
                    <option key={v.id} value={v.label} />
                  ))}
                </datalist>
              </div>
              <button
                type="button"
                onClick={() => {
                  const first = variantOptions[0]
                  if (first && !fields.some((f) => f.variantId === first.id)) {
                    addVariant(first.id)
                  } else {
                    append({
                      variantId: '',
                      variantLabel: '',
                      quantityOrdered: '1',
                      costPrice: '0',
                    })
                  }
                }}
                className="inline-flex h-[38px] items-center rounded-md border px-3.5 text-[13px] font-medium"
                style={{ borderColor: BORDER_STRONG, color: FG, background: 'white' }}
              >
                {t('inventory.browse', 'Browse')}
              </button>
            </div>

            {fields.length === 0 ? (
              <div
                className="mt-4 flex flex-col items-center gap-1.5 rounded-[10px] px-4 py-6 text-center"
                style={{
                  border: `1px dashed ${BORDER_STRONG}`,
                  background: `${MUTED} / 0.4`,
                }}
              >
                <div
                  className="mb-1 grid h-10 w-10 place-items-center rounded-[10px]"
                  style={{
                    background: 'white',
                    border: `1px solid ${BORDER}`,
                    color: MUTED_FG,
                  }}
                >
                  <Package className="h-5 w-5" strokeWidth={1.5} />
                </div>
                <div className="text-[13px] font-semibold" style={{ color: FG }}>
                  {t('inventory.noItemsTitle', 'No products added yet')}
                </div>
                <div className="text-[12.5px]" style={{ color: MUTED_FG }}>
                  {t(
                    'inventory.noItemsHint',
                    'Search or browse your catalog to add items to this order.',
                  )}
                </div>
              </div>
            ) : (
              <div className="mt-[18px]" style={{ borderTop: `1px solid ${BORDER}` }}>
                <div
                  className="grid items-center gap-3 pb-2.5 pt-3.5 text-[12.5px] font-medium"
                  style={{
                    gridTemplateColumns: 'minmax(0,2fr) minmax(0,1fr) 84px 110px 90px 28px',
                    color: MUTED_FG,
                  }}
                >
                  <div>{t('inventory.products', 'Products')}</div>
                  <div>{t('inventory.supplierSku', 'Supplier SKU')}</div>
                  <div className="text-right">{t('inventory.qty', 'Quantity')}</div>
                  <div className="text-right">{t('inventory.unitCost', 'Cost')}</div>
                  <div className="text-right">{t('inventory.lineTotal', 'Total')}</div>
                  <div></div>
                </div>
                {fields.map((f, i) => {
                  const qty = Number(form.watch(`items.${i}.quantityOrdered`) || '0')
                  const unit = Number(form.watch(`items.${i}.costPrice`) || '0')
                  const line = Number.isNaN(qty) || Number.isNaN(unit) ? 0 : qty * unit
                  const variantId = form.watch(`items.${i}.variantId`)
                  const v = variantOptions.find((x) => x.id === variantId)
                  return (
                    <div
                      key={f.id}
                      className="grid items-center gap-3 py-3.5"
                      style={{
                        gridTemplateColumns:
                          'minmax(0,2fr) minmax(0,1fr) 84px 110px 90px 28px',
                        borderTop: `1px solid ${BORDER}`,
                      }}
                    >
                      <div className="flex min-w-0 items-center gap-2.5">
                        <div
                          className="grid h-8 w-8 flex-shrink-0 place-items-center rounded-md"
                          style={{
                            background: MUTED,
                            border: `1px solid ${BORDER}`,
                            color: MUTED_FG,
                          }}
                        >
                          <Package className="h-4 w-4" strokeWidth={1.8} />
                        </div>
                        <select
                          value={variantId}
                          onChange={(e) => {
                            const picked = variantOptions.find(
                              (x) => x.id === e.target.value,
                            )
                            form.setValue(`items.${i}.variantId`, e.target.value)
                            form.setValue(
                              `items.${i}.variantLabel`,
                              picked?.label ?? '',
                            )
                            if (picked && !form.getValues(`items.${i}.costPrice`)) {
                              form.setValue(`items.${i}.costPrice`, picked.costPrice)
                            }
                          }}
                          className="h-[34px] min-w-0 flex-1 rounded-[7px] border bg-white px-2.5 text-[13px] font-medium"
                          style={{ borderColor: BORDER_STRONG, color: PRIMARY }}
                        >
                          <option value="">{t('inventory.selectVariant', 'Select product…')}</option>
                          {variantOptions.map((opt) => (
                            <option key={opt.id} value={opt.id}>
                              {opt.label}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div
                        className="truncate text-[12px]"
                        style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                      >
                        {v?.sku || '—'}
                      </div>
                      <div>
                        <input
                          type="number"
                          step="0.0001"
                          {...form.register(`items.${i}.quantityOrdered`)}
                          className="h-[34px] w-full rounded-[7px] border px-2.5 text-right text-[13px]"
                          style={{
                            borderColor: BORDER_STRONG,
                            background: 'white',
                            color: FG,
                            fontVariantNumeric: 'tabular-nums',
                          }}
                        />
                      </div>
                      <div className="relative">
                        <span
                          className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-[12.5px]"
                          style={{ color: MUTED_FG }}
                        >
                          $
                        </span>
                        <input
                          type="number"
                          step="0.0001"
                          {...form.register(`items.${i}.costPrice`)}
                          className="h-[34px] w-full rounded-[7px] border pl-6 pr-2.5 text-right text-[13px]"
                          style={{
                            borderColor: BORDER_STRONG,
                            background: 'white',
                            color: FG,
                            fontVariantNumeric: 'tabular-nums',
                          }}
                        />
                      </div>
                      <div
                        className="text-right text-[13.5px] font-medium"
                        style={{ color: FG, fontVariantNumeric: 'tabular-nums' }}
                      >
                        ${line.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </div>
                      <button
                        type="button"
                        onClick={() => remove(i)}
                        className="grid h-7 w-7 place-items-center rounded-md"
                        style={{ color: MUTED_FG }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.background = MUTED
                          e.currentTarget.style.color = DESTRUCTIVE
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.background = 'transparent'
                          e.currentTarget.style.color = MUTED_FG
                        }}
                        aria-label={t('common.remove', 'Remove')}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  )
                })}
                <div
                  className="pb-0.5 pt-3.5 text-[12.5px]"
                  style={{ borderTop: `1px solid ${BORDER}`, color: MUTED_FG }}
                >
                  {t('inventory.variantsOnOrder', {
                    count: variantCount,
                    defaultValue: '{{count}} variants on purchase order',
                  })}
                </div>
              </div>
            )}
          </Card>

          <div className="grid gap-[14px] md:grid-cols-2">
            <Card>
              <CardTitle>{t('inventory.additionalDetails', 'Additional details')}</CardTitle>
              <Field label={t('inventory.referenceNumber', 'Reference number')}>
                <input
                  type="text"
                  value={refNumber}
                  onChange={(e) => setRefNumber(e.target.value)}
                  className={inputCls}
                  style={inputStyle}
                />
              </Field>
              <Field label={t('inventory.noteToSupplier', 'Note to supplier')}>
                <div className="relative">
                  <textarea
                    rows={3}
                    maxLength={5000}
                    {...form.register('notes')}
                    className="w-full resize-y rounded-[8px] border px-3 py-2.5 text-[13.5px] leading-[1.5]"
                    style={{
                      borderColor: BORDER_STRONG,
                      background: 'white',
                      color: FG,
                      minHeight: 92,
                    }}
                  />
                  <span
                    className="pointer-events-none absolute bottom-2 right-2.5 rounded-[3px] bg-white px-1 text-[11.5px]"
                    style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                  >
                    {(form.watch('notes') ?? '').length}/5000
                  </span>
                </div>
              </Field>
            </Card>

            <Card>
              <div className="mb-3.5 flex items-baseline gap-2.5">
                <div className="text-[13.5px] font-semibold" style={{ color: FG }}>
                  {t('inventory.costSummary', 'Cost summary')}
                </div>
              </div>
              <CostRow label={t('inventory.taxesIncluded', 'Taxes (Included)')}>
                $0.00
              </CostRow>
              <CostRow label={t('inventory.subtotal', 'Subtotal')} strong>
                ${subtotal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </CostRow>
              <div className="pb-2 text-[12px]" style={{ color: MUTED_FG }}>
                {t('inventory.itemsCount', {
                  count: itemCount,
                  defaultValue: '{{count}} items',
                })}
              </div>

              <div className="mt-3.5 mb-1 text-[13.5px] font-semibold" style={{ color: FG }}>
                {t('inventory.costAdjustments', 'Cost adjustments')}
              </div>
              <div
                className="flex items-center justify-between py-[5px] text-[13px]"
              >
                <span style={{ color: FG }}>{t('inventory.shipping', 'Shipping')}</span>
                <div className="relative w-24">
                  <span
                    className="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-[12px]"
                    style={{ color: MUTED_FG }}
                  >
                    $
                  </span>
                  <input
                    type="number"
                    step="0.01"
                    value={shippingCost}
                    onChange={(e) => setShippingCost(e.target.value)}
                    className="h-8 w-full rounded-[6px] border pl-5 pr-2 text-right text-[13px]"
                    style={{
                      borderColor: BORDER_STRONG,
                      background: 'white',
                      color: FG,
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  />
                </div>
              </div>

              <div className="my-2.5 h-px" style={{ background: BORDER }} />
              <div className="flex items-baseline justify-between pt-1.5 text-[13px]">
                <span className="font-semibold" style={{ color: FG }}>
                  {t('inventory.total', 'Total')}
                </span>
                <span
                  className="font-semibold"
                  style={{ color: FG, fontVariantNumeric: 'tabular-nums' }}
                >
                  ${total.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </span>
              </div>
            </Card>
          </div>
        </fieldset>

        <div className="flex items-center justify-between gap-3 pt-1">
          <button
            type="button"
            onClick={() => void backToList()}
            className="inline-flex h-9 items-center rounded-md border px-3.5 text-[13px] font-medium"
            style={{ borderColor: BORDER, background: 'white', color: FG }}
            disabled={submitting}
          >
            {t('common.discard', 'Discard')}
          </button>
          {!readOnly && (
            <div className="flex gap-2">
              <button
                type="submit"
                className="inline-flex h-9 items-center rounded-md border px-3.5 text-[13px] font-medium"
                style={{ borderColor: BORDER, background: 'white', color: FG }}
                disabled={submitting}
              >
                {submitting ? t('common.saving') : t('inventory.saveDraft', 'Save as draft')}
              </button>
              <button
                type="submit"
                className="inline-flex h-9 items-center rounded-md px-3.5 text-[13px] font-medium text-white"
                style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
                disabled={submitting}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
                }}
              >
                {submitting
                  ? t('common.saving')
                  : existing
                    ? t('common.save')
                    : t('inventory.createPurchaseOrder', 'Create purchase order')}
              </button>
            </div>
          )}
        </div>
      </form>
    </div>
  )
}

const inputCls =
  'h-[38px] w-full rounded-[8px] border bg-white px-3 text-[13.5px] outline-none transition-[border-color,box-shadow] duration-150 focus:shadow-[0_0_0_3px_hsl(221_83%_53%/0.15)]'
const inputStyle: React.CSSProperties = {
  borderColor: BORDER_STRONG,
  color: FG,
}

function Card({ children }: { children: React.ReactNode }) {
  return (
    <section
      className="rounded-[12px] bg-white px-5 py-[18px]"
      style={{
        border: `1px solid ${BORDER}`,
        boxShadow: '0 1px 2px rgba(16,24,40,0.04)',
      }}
    >
      {children}
    </section>
  )
}

function CardTitle({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="mb-3.5 text-[13.5px] font-semibold tracking-[-0.005em]"
      style={{ color: FG }}
    >
      {children}
    </div>
  )
}

function Grid({
  cols,
  children,
}: {
  cols: 2 | 3
  children: React.ReactNode
}) {
  return (
    <div
      className="grid gap-x-5 gap-y-4"
      style={{
        gridTemplateColumns:
          cols === 3
            ? 'repeat(3, minmax(0,1fr))'
            : 'repeat(2, minmax(0,1fr))',
      }}
    >
      {children}
    </div>
  )
}

function Field({
  label,
  optional,
  children,
}: {
  label: string
  optional?: boolean
  children: React.ReactNode
}) {
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <label className="text-[12.5px] font-medium" style={{ color: FG }}>
        {label}{' '}
        {optional && (
          <span className="font-normal" style={{ color: MUTED_FG }}>
            (optional)
          </span>
        )}
      </label>
      {children}
    </div>
  )
}

function Divider() {
  return <div className="-mx-5 my-[18px] h-px" style={{ background: BORDER }} />
}

function NativeSelect({
  value,
  onChange,
  children,
}: {
  value: string
  onChange: (v: string) => void
  children: React.ReactNode
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={`${inputCls} cursor-pointer pr-9 appearance-none`}
        style={inputStyle}
      >
        {children}
      </select>
      <ChevronDown
        className="pointer-events-none absolute right-2.5 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
        style={{ color: MUTED_FG }}
      />
    </div>
  )
}

function CostRow({
  label,
  strong,
  children,
}: {
  label: string
  strong?: boolean
  children: React.ReactNode
}) {
  return (
    <div className="flex items-baseline justify-between py-[5px] text-[13px]">
      <span style={{ color: FG, fontWeight: strong ? 600 : 400 }}>{label}</span>
      <span
        style={{
          color: FG,
          fontVariantNumeric: 'tabular-nums',
          fontWeight: strong ? 600 : 400,
        }}
      >
        {children}
      </span>
    </div>
  )
}
