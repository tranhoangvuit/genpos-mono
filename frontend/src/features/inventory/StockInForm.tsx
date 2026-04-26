import { useNavigate } from '@tanstack/react-router'
import {
  ArrowRight,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Package,
  Search,
  Trash2,
  X,
} from 'lucide-react'
import { forwardRef, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'
import { currencySymbol, formatMoney } from '@/shared/lib/currency'

import { useStores, useVariantPicker } from './hooks'
import { ProductPickerDialog } from './ProductPickerDialog'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const BORDER_STRONG = 'hsl(214 32% 85%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const PRIMARY = 'hsl(221 83% 53%)'
const DESTRUCTIVE = 'hsl(0 84% 60%)'

type LineItem = {
  variantId: string
  productName: string
  variantName: string
  sku: string
  quantity: string
  unitCost: string
  adjustReason: string
  adjustNote: string
}

const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const RED_SOFT = 'hsl(0 93% 94%)'
const RED_INK = 'hsl(0 74% 42%)'

const ADJUST_REASONS: Array<{ value: string; label: string }> = [
  { value: 'stock_count', label: 'Stock count' },
  { value: 'received', label: 'Received' },
  { value: 'damaged', label: 'Damaged' },
  { value: 'theft_loss', label: 'Theft / loss' },
  { value: 'return', label: 'Return' },
  { value: 'other', label: 'Other' },
]

const REASONS: Array<{ value: string; label: string }> = [
  { value: 'receive', label: 'Receive without PO' },
  { value: 'initial', label: 'Initial stock' },
  { value: 'transfer', label: 'Transfer in' },
  { value: 'found', label: 'Found / Recount' },
  { value: 'return', label: 'Customer return' },
  { value: 'other', label: 'Other' },
]

function todayString() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function StockInForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')

  const { data: stores } = useStores()
  const { data: variants } = useVariantPicker()
  const currency = 'VND'

  const [reason, setReason] = useState('receive')
  const [referenceNumber, setReferenceNumber] = useState('')
  const [storeId, setStoreId] = useState('')
  const [dateReceived, setDateReceived] = useState(todayString())
  const [note, setNote] = useState('')
  const [items, setItems] = useState<LineItem[]>([])

  const [prodSearch, setProdSearch] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)
  const [pickerSeed, setPickerSeed] = useState('')

  const [adjustPopover, setAdjustPopover] = useState<
    { index: number; prevQuantity: string } | null
  >(null)
  const popoverRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!adjustPopover) return
    function onDown(e: MouseEvent) {
      if (popoverRef.current?.contains(e.target as Node)) return
      const target = e.target as HTMLElement
      if (target.closest('[data-qty-input]')) return
      cancelAdjust()
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') cancelAdjust()
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [adjustPopover])

  function cancelAdjust() {
    if (!adjustPopover) return
    const { index, prevQuantity } = adjustPopover
    setItems((prev) =>
      prev.map((it, idx) =>
        idx === index ? { ...it, quantity: prevQuantity } : it,
      ),
    )
    setAdjustPopover(null)
  }

  function saveAdjust() {
    setAdjustPopover(null)
  }

  useEffect(() => {
    if (!storeId && stores && stores.length > 0) {
      setStoreId(stores[0].id)
    }
  }, [storeId, stores])

  const subtotalQty = items.reduce((s, it) => s + (Number(it.quantity) || 0), 0)
  const totalCost = items.reduce(
    (s, it) => s + (Number(it.quantity) || 0) * (Number(it.unitCost) || 0),
    0,
  )
  const variantCount = items.length

  const openPicker = (seed = '') => {
    setPickerSeed(seed)
    setPickerOpen(true)
  }

  const addVariants = (ids: string[]) => {
    setItems((prev) => {
      const existing = new Set(prev.map((p) => p.variantId))
      const next = [...prev]
      for (const id of ids) {
        if (existing.has(id)) continue
        const v = variants?.find((x) => x.id === id)
        if (!v) continue
        next.push({
          variantId: v.id,
          productName: v.productName,
          variantName: v.variantName,
          sku: v.sku,
          quantity: '1',
          unitCost: v.costPrice || '0',
          adjustReason: '',
          adjustNote: '',
        })
      }
      return next
    })
    setProdSearch('')
  }

  const removeItem = (i: number) => {
    setItems((prev) => prev.filter((_, idx) => idx !== i))
  }

  const updateItem = (i: number, patch: Partial<LineItem>) => {
    setItems((prev) => prev.map((it, idx) => (idx === i ? { ...it, ...patch } : it)))
  }

  const backToList = () =>
    navigate({
      to: '/$subdomain/inventory/stock-ins',
      params: { subdomain },
    })

  // Stock-in submission requires backend RPC support (not yet exposed).
  // Until then, persist nothing and return to the list.
  const onSave = () => void backToList()

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
          className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]"
          style={{ color: FG }}
        >
          {t('inventory.newStockIn', 'New stock in')}
          <span
            className="rounded-md px-2 py-[2px] text-[12px] font-medium"
            style={{ background: MUTED, color: MUTED_FG }}
          >
            {t('inventory.noPurchaseOrder', 'No purchase order')}
          </span>
        </h1>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          onSave()
        }}
        noValidate
        className="flex flex-col gap-[14px] pb-20"
      >
        <Card>
          <Grid cols={2}>
            <Field label={t('inventory.reason', 'Reason')}>
              <NativeSelect value={reason} onChange={setReason}>
                {REASONS.map((r) => (
                  <option key={r.value} value={r.value}>
                    {r.label}
                  </option>
                ))}
              </NativeSelect>
            </Field>
            <Field label={t('inventory.referenceNumber', 'Reference number')} optional>
              <input
                type="text"
                value={referenceNumber}
                onChange={(e) => setReferenceNumber(e.target.value)}
                placeholder="e.g. INV-2026-0418"
                className={inputCls}
                style={inputStyle}
              />
            </Field>
          </Grid>
          <Divider />
          <Grid cols={2}>
            <Field label={t('inventory.destination', 'Destination')}>
              <NativeSelect value={storeId} onChange={setStoreId}>
                <option value="">—</option>
                {(stores ?? []).map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </NativeSelect>
            </Field>
            <Field label={t('inventory.dateReceived', 'Date received')}>
              <div className="relative">
                <input
                  type="date"
                  value={dateReceived}
                  onChange={(e) => setDateReceived(e.target.value)}
                  className={`${inputCls} pr-9`}
                  style={inputStyle}
                />
                <CalendarDays
                  className="pointer-events-none absolute right-3 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
                  style={{ color: MUTED_FG }}
                />
              </div>
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
                placeholder={t('inventory.searchProducts', 'Search products')}
                value={prodSearch}
                onChange={(e) => {
                  const v = e.target.value
                  setProdSearch(v)
                  if (v && !pickerOpen) openPicker(v)
                }}
                onFocus={() => {
                  if (prodSearch && !pickerOpen) openPicker(prodSearch)
                }}
                className={`${inputCls} pl-9`}
                style={inputStyle}
                autoComplete="off"
              />
            </div>
            <button
              type="button"
              onClick={() => openPicker(prodSearch)}
              className="inline-flex h-[38px] items-center rounded-md border px-3.5 text-[13px] font-medium"
              style={{ borderColor: BORDER_STRONG, color: FG, background: 'white' }}
            >
              {t('inventory.browse', 'Browse')}
            </button>
          </div>

          {items.length === 0 ? (
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
                  'inventory.stockInEmptyHint',
                  'Search or browse your catalog to record received stock.',
                )}
              </div>
            </div>
          ) : (
            <div className="mt-[18px]" style={{ borderTop: `1px solid ${BORDER}` }}>
              <div
                className="grid items-center gap-3 pb-2.5 pt-3.5 text-[12.5px] font-medium"
                style={{
                  gridTemplateColumns: 'minmax(0,2fr) minmax(0,1fr) 96px 110px 90px 28px',
                  color: MUTED_FG,
                }}
              >
                <div>{t('inventory.products', 'Products')}</div>
                <div>{t('inventory.sku', 'SKU')}</div>
                <div className="text-right">{t('inventory.qtyReceived', 'Qty received')}</div>
                <div className="text-right">{t('inventory.unitCost', 'Unit cost')}</div>
                <div className="text-right">{t('inventory.lineTotal', 'Total')}</div>
                <div></div>
              </div>
              {items.map((it, i) => {
                const qty = Number(it.quantity) || 0
                const unit = Number(it.unitCost) || 0
                const line = qty * unit
                const popoverOpen = adjustPopover?.index === i
                return (
                  <div
                    key={it.variantId}
                    className="relative grid items-center gap-3 py-3.5"
                    style={{
                      gridTemplateColumns:
                        'minmax(0,2fr) minmax(0,1fr) 96px 110px 90px 28px',
                      borderTop: `1px solid ${BORDER}`,
                    }}
                  >
                    {popoverOpen && (
                      <AdjustPopover
                        ref={popoverRef}
                        productLabel={
                          it.variantName && it.variantName !== it.productName
                            ? `${it.productName} · ${it.variantName}`
                            : it.productName
                        }
                        fromQty={Number(adjustPopover?.prevQuantity ?? '0') || 0}
                        toQty={qty}
                        reason={it.adjustReason}
                        note={it.adjustNote}
                        onReason={(r) => updateItem(i, { adjustReason: r })}
                        onNote={(n) => updateItem(i, { adjustNote: n })}
                        onCancel={cancelAdjust}
                        onSave={saveAdjust}
                      />
                    )}
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
                      <div className="flex min-w-0 flex-col gap-0.5">
                        <span
                          className="truncate text-[13.5px] font-medium"
                          style={{ color: PRIMARY }}
                        >
                          {it.productName}
                        </span>
                        {it.variantName && it.variantName !== it.productName && (
                          <span
                            className="truncate text-[12px]"
                            style={{ color: MUTED_FG }}
                          >
                            {it.variantName}
                          </span>
                        )}
                      </div>
                    </div>
                    <div
                      className="truncate text-[12.5px]"
                      style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                    >
                      {it.sku || '—'}
                    </div>
                    <div>
                      <input
                        type="number"
                        step="1"
                        min="0"
                        value={it.quantity}
                        data-qty-input
                        onFocus={() => {
                          if (adjustPopover?.index !== i) {
                            setAdjustPopover({
                              index: i,
                              prevQuantity: it.quantity,
                            })
                          }
                        }}
                        onChange={(e) => {
                          const value = e.target.value
                          if (adjustPopover?.index !== i) {
                            setAdjustPopover({
                              index: i,
                              prevQuantity: it.quantity,
                            })
                          }
                          updateItem(i, { quantity: value })
                        }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault()
                            saveAdjust()
                          }
                        }}
                        className="h-[34px] w-full rounded-[7px] border px-2.5 text-right text-[13px] transition-colors"
                        style={{
                          borderColor: popoverOpen ? PRIMARY : BORDER_STRONG,
                          boxShadow: popoverOpen
                            ? `0 0 0 3px hsl(221 83% 53% / 0.15)`
                            : 'none',
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
                        {currencySymbol(currency)}
                      </span>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={it.unitCost}
                        onChange={(e) => updateItem(i, { unitCost: e.target.value })}
                        className="h-[34px] w-full rounded-[7px] border pl-7 pr-2.5 text-right text-[13px]"
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
                      {formatMoney(line, currency)}
                    </div>
                    <button
                      type="button"
                      onClick={() => removeItem(i)}
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
                {t('inventory.variantsInStockIn', {
                  count: variantCount,
                  defaultValue: '{{count}} variants in this stock in',
                })}
              </div>
            </div>
          )}
        </Card>

        <div className="grid gap-[14px] md:grid-cols-2">
          <Card>
            <CardTitle>{t('inventory.additionalDetails', 'Additional details')}</CardTitle>
            <Field label={t('inventory.noteToReceiver', 'Note to receiver')} optional>
              <div className="relative">
                <textarea
                  rows={3}
                  maxLength={5000}
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="e.g. 2 cases damaged in transit — set aside for review"
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
                  {note.length}/5000
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
            <div className="flex items-baseline justify-between py-[5px] text-[13px]">
              <span className="font-semibold" style={{ color: FG }}>
                {t('inventory.itemsReceived', 'Items received')}
              </span>
              <span
                className="font-semibold"
                style={{ color: FG, fontVariantNumeric: 'tabular-nums' }}
              >
                {subtotalQty.toLocaleString()}
              </span>
            </div>
            <div className="pb-2 text-[12px]" style={{ color: MUTED_FG }}>
              {t('inventory.variantsCount', {
                count: variantCount,
                defaultValue: '{{count}} variants',
              })}
            </div>

            <div className="my-2.5 h-px" style={{ background: BORDER }} />
            <div className="flex items-baseline justify-between pt-1.5 text-[13px]">
              <span className="font-semibold" style={{ color: FG }}>
                {t('inventory.totalCost', 'Total cost')}
              </span>
              <span
                className="font-semibold"
                style={{ color: FG, fontVariantNumeric: 'tabular-nums' }}
              >
                {formatMoney(totalCost, currency)}
              </span>
            </div>
            <div className="pt-1 text-[12px]" style={{ color: MUTED_FG }}>
              {t(
                'inventory.stockInValuation',
                'Inventory valuation will increase by this amount.',
              )}
            </div>
          </Card>
        </div>

        <div className="flex items-center justify-between gap-3 pt-1">
          <button
            type="button"
            onClick={() => void backToList()}
            className="inline-flex h-9 items-center rounded-md border px-3.5 text-[13px] font-medium"
            style={{ borderColor: BORDER, background: 'white', color: FG }}
          >
            {t('common.discard', 'Discard')}
          </button>
          <div className="flex gap-2">
            <button
              type="submit"
              className="inline-flex h-9 items-center rounded-md border px-3.5 text-[13px] font-medium"
              style={{ borderColor: BORDER, background: 'white', color: FG }}
            >
              {t('inventory.saveDraft', 'Save as draft')}
            </button>
            <button
              type="submit"
              disabled={items.length === 0}
              className="inline-flex h-9 items-center rounded-md px-3.5 text-[13px] font-medium text-white disabled:cursor-not-allowed disabled:opacity-60"
              style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
              onMouseEnter={(e) => {
                if (items.length === 0) return
                e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
              }}
            >
              {t('inventory.receiveStock', 'Receive stock')}
            </button>
          </div>
        </div>
      </form>

      <ProductPickerDialog
        open={pickerOpen}
        onOpenChange={(o) => {
          setPickerOpen(o)
          if (!o) setProdSearch('')
        }}
        variants={variants ?? []}
        excludeVariantIds={items.map((i) => i.variantId)}
        initialQuery={pickerSeed}
        onAdd={addVariants}
      />
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

function Grid({ cols, children }: { cols: 2 | 3; children: React.ReactNode }) {
  return (
    <div
      className="grid gap-x-5 gap-y-4"
      style={{
        gridTemplateColumns:
          cols === 3 ? 'repeat(3, minmax(0,1fr))' : 'repeat(2, minmax(0,1fr))',
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
  return (
    <div className="-mx-5 my-[18px] h-px" style={{ background: BORDER }} />
  )
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
        className={`${inputCls} appearance-none pr-8`}
        style={inputStyle}
      >
        {children}
      </select>
      <ChevronRight
        className="pointer-events-none absolute right-2 top-1/2 h-[14px] w-[14px] -translate-y-1/2 rotate-90"
        style={{ color: MUTED_FG }}
      />
    </div>
  )
}

type AdjustPopoverProps = {
  productLabel: string
  fromQty: number
  toQty: number
  reason: string
  note: string
  onReason: (r: string) => void
  onNote: (n: string) => void
  onCancel: () => void
  onSave: () => void
}

const AdjustPopover = forwardRef<HTMLDivElement, AdjustPopoverProps>(
  function AdjustPopover(
    { productLabel, fromQty, toQty, reason, note, onReason, onNote, onCancel, onSave },
    ref,
  ) {
    const delta = toQty - fromQty
    const deltaSign = delta > 0 ? '+' : ''
    const deltaStyle =
      delta > 0
        ? { background: DONE_SOFT, color: DONE_INK }
        : delta < 0
          ? { background: RED_SOFT, color: RED_INK }
          : { background: MUTED, color: FG }
    return (
      <div
        ref={ref}
        role="dialog"
        aria-label="Adjustment reason"
        className="absolute z-30 w-[340px] rounded-[8px] bg-white p-3.5"
        style={{
          right: 'calc(100% + 12px)',
          top: '50%',
          transform: 'translateY(-50%)',
          border: `1px solid ${BORDER}`,
          boxShadow: '0 12px 40px hsl(222 47% 11% / 0.14)',
        }}
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="mb-2.5 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="text-[13px] font-semibold" style={{ color: FG }}>
              Adjust quantity
            </div>
            <div className="mt-0.5 truncate text-[12px]" style={{ color: MUTED_FG }}>
              {productLabel}
            </div>
          </div>
          <button
            type="button"
            onClick={onCancel}
            className="grid h-6 w-6 place-items-center rounded transition"
            style={{ color: MUTED_FG }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = MUTED
              e.currentTarget.style.color = FG
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'transparent'
              e.currentTarget.style.color = MUTED_FG
            }}
            aria-label="Close"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>

        <div
          className="mb-3 flex items-center justify-between rounded-[6px] px-3 py-2.5"
          style={{ background: `${MUTED}` }}
        >
          <div
            className="inline-flex items-center gap-2 text-[14px] font-medium"
            style={{ color: FG, fontVariantNumeric: 'tabular-nums' }}
          >
            <span>{fromQty.toLocaleString()}</span>
            <ArrowRight
              className="h-3.5 w-3.5"
              style={{ color: MUTED_FG }}
              strokeWidth={2.2}
            />
            <span>{toQty.toLocaleString()}</span>
          </div>
          <span
            className="rounded-full px-2 py-[2px] text-[12.5px] font-semibold"
            style={{ ...deltaStyle, fontVariantNumeric: 'tabular-nums' }}
          >
            {delta === 0 ? '0' : `${deltaSign}${delta}`}
          </span>
        </div>

        <div
          className="mb-1.5 text-[11px] font-semibold uppercase tracking-[0.06em]"
          style={{ color: MUTED_FG }}
        >
          Reason
        </div>
        <div className="mb-3 flex flex-wrap gap-1.5">
          {ADJUST_REASONS.map((r) => {
            const active = reason === r.value
            return (
              <button
                key={r.value}
                type="button"
                onClick={() => onReason(r.value)}
                className="rounded-full border px-2.5 py-[5px] text-[12px] font-medium transition-colors"
                style={{
                  borderColor: active ? FG : BORDER,
                  background: active ? FG : 'white',
                  color: active ? 'white' : FG,
                }}
              >
                {r.label}
              </button>
            )
          })}
        </div>

        <div
          className="mb-1.5 text-[11px] font-semibold uppercase tracking-[0.06em]"
          style={{ color: MUTED_FG }}
        >
          Note <span className="font-normal normal-case tracking-normal">(optional)</span>
        </div>
        <textarea
          rows={2}
          value={note}
          onChange={(e) => onNote(e.target.value)}
          placeholder="Add context for this change…"
          className="mb-3 w-full resize-y rounded-[6px] border px-2.5 py-2 text-[13px] outline-none focus:shadow-[0_0_0_2px_hsl(221_83%_53%/0.15)]"
          style={{
            borderColor: BORDER_STRONG,
            color: FG,
            background: 'white',
            minHeight: 56,
          }}
        />

        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            className="inline-flex h-8 items-center rounded-md border px-3 text-[12.5px] font-medium"
            style={{ borderColor: BORDER, color: FG, background: 'white' }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onSave}
            className="inline-flex h-8 items-center rounded-md px-3 text-[12.5px] font-medium text-white"
            style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
          >
            Save adjustment
          </button>
        </div>
      </div>
    )
  },
)
