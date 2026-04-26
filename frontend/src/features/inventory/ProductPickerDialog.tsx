import * as DialogPrimitive from '@radix-ui/react-dialog'
import { ChevronDown, Package, Plus, Search, X } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/shared/lib/cn'
import { DialogOverlay, DialogPortal } from '@/shared/ui/dialog'

import type { VariantPickerRow } from './types'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const BORDER_STRONG = 'hsl(214 32% 85%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const PRIMARY = 'hsl(221 83% 53%)'
const PRIMARY_INK = 'hsl(224 76% 48%)'

type SearchBy = 'all' | 'title' | 'sku'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  variants: VariantPickerRow[]
  excludeVariantIds?: string[]
  initialQuery?: string
  onAdd: (variantIds: string[]) => void
}

type ProductGroup = {
  productName: string
  variants: VariantPickerRow[]
}

export function ProductPickerDialog({
  open,
  onOpenChange,
  variants,
  excludeVariantIds = [],
  initialQuery = '',
  onAdd,
}: Props) {
  const { t } = useTranslation()
  const [query, setQuery] = useState(initialQuery)
  const [searchBy, setSearchBy] = useState<SearchBy>('all')
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const searchRef = useRef<HTMLInputElement>(null)
  const excluded = useMemo(() => new Set(excludeVariantIds), [excludeVariantIds])

  useEffect(() => {
    if (!open) return
    setQuery(initialQuery)
    setSelected(new Set())
    const id = window.setTimeout(() => searchRef.current?.focus(), 50)
    return () => window.clearTimeout(id)
  }, [open, initialQuery])

  const groups = useMemo<ProductGroup[]>(() => {
    const map = new Map<string, ProductGroup>()
    for (const v of variants) {
      if (excluded.has(v.id)) continue
      const key = v.productName || '—'
      let g = map.get(key)
      if (!g) {
        g = { productName: key, variants: [] }
        map.set(key, g)
      }
      g.variants.push(v)
    }
    return Array.from(map.values()).sort((a, b) =>
      a.productName.localeCompare(b.productName),
    )
  }, [variants, excluded])

  const filteredGroups = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return groups
    return groups
      .map((g) => {
        const matchedProduct =
          searchBy !== 'sku' && g.productName.toLowerCase().includes(q)
        const variants = g.variants.filter((v) => {
          if (searchBy === 'title') return v.variantName.toLowerCase().includes(q)
          if (searchBy === 'sku') return (v.sku || '').toLowerCase().includes(q)
          return (
            v.variantName.toLowerCase().includes(q) ||
            (v.sku || '').toLowerCase().includes(q)
          )
        })
        if (matchedProduct) return g
        if (variants.length === 0) return null
        return { ...g, variants }
      })
      .filter((g): g is ProductGroup => g !== null)
  }, [groups, query, searchBy])

  function toggleVariant(id: string) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleProduct(g: ProductGroup) {
    setSelected((prev) => {
      const next = new Set(prev)
      const ids = g.variants.map((v) => v.id)
      const allSelected = ids.every((id) => next.has(id))
      if (allSelected) ids.forEach((id) => next.delete(id))
      else ids.forEach((id) => next.add(id))
      return next
    })
  }

  function productState(g: ProductGroup): 'all' | 'some' | 'none' {
    const ids = g.variants.map((v) => v.id)
    const sel = ids.filter((id) => selected.has(id))
    if (sel.length === 0) return 'none'
    if (sel.length === ids.length) return 'all'
    return 'some'
  }

  const count = selected.size
  const canAdd = count > 0

  function handleAdd() {
    if (!canAdd) return
    onAdd(Array.from(selected))
    onOpenChange(false)
  }

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        <DialogOverlay className="bg-[hsl(222_47%_11%/0.35)] backdrop-blur-[2px]" />
        <DialogPrimitive.Content
          aria-describedby={undefined}
          className={cn(
            'fixed left-[50%] top-[8vh] z-50 flex w-[calc(100%-40px)] max-w-[780px] -translate-x-1/2 flex-col overflow-hidden rounded-[14px] bg-white shadow-2xl',
          )}
          style={{ maxHeight: 'calc(100vh - 16vh)' }}
        >
          <div
            className="flex items-center justify-between px-5 py-4"
            style={{ background: `${MUTED}`, borderBottom: `1px solid ${BORDER}` }}
          >
            <DialogPrimitive.Title className="m-0 text-[15px] font-semibold tracking-[-0.005em]">
              {t('inventory.addProducts', 'Add products')}
            </DialogPrimitive.Title>
            <DialogPrimitive.Close
              className="grid h-7 w-7 place-items-center rounded-md transition"
              style={{ color: MUTED_FG }}
              aria-label={t('common.close', 'Close')}
            >
              <X className="h-4 w-4" />
            </DialogPrimitive.Close>
          </div>

          <div className="grid grid-cols-[1fr_200px] gap-2.5 px-5 pb-2.5 pt-3.5">
            <div className="relative">
              <Search
                className="pointer-events-none absolute left-3 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
                style={{ color: MUTED_FG }}
              />
              <input
                ref={searchRef}
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t('inventory.searchProducts', 'Search products')}
                className="h-[38px] w-full rounded-[8px] border bg-white pl-9 pr-9 text-[13.5px] outline-none transition-[border-color,box-shadow] duration-150 focus:shadow-[0_0_0_3px_hsl(221_83%_53%/0.15)]"
                style={{ borderColor: BORDER_STRONG, color: FG }}
              />
              {query && (
                <button
                  type="button"
                  onClick={() => {
                    setQuery('')
                    searchRef.current?.focus()
                  }}
                  className="absolute right-2 top-1/2 grid h-[22px] w-[22px] -translate-y-1/2 place-items-center rounded-full"
                  style={{ background: MUTED, color: MUTED_FG }}
                  aria-label={t('common.clear', 'Clear')}
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
            <div className="relative">
              <select
                value={searchBy}
                onChange={(e) => setSearchBy(e.target.value as SearchBy)}
                className="h-[38px] w-full appearance-none rounded-[8px] border bg-white pl-3 pr-9 text-[13.5px] outline-none focus:shadow-[0_0_0_3px_hsl(221_83%_53%/0.15)]"
                style={{ borderColor: BORDER_STRONG, color: FG }}
              >
                <option value="all">{t('inventory.searchByAll', 'Search by All')}</option>
                <option value="title">{t('inventory.searchByTitle', 'Search by Title')}</option>
                <option value="sku">{t('inventory.searchBySku', 'Search by SKU')}</option>
              </select>
              <ChevronDown
                className="pointer-events-none absolute right-2.5 top-1/2 h-[14px] w-[14px] -translate-y-1/2"
                style={{ color: MUTED_FG }}
              />
            </div>
          </div>

          <div className="flex items-center gap-2 px-5 pb-3">
            <button
              type="button"
              className="inline-flex h-7 items-center gap-1 rounded-full border border-dashed bg-white px-2.5 text-[12.5px] font-medium"
              style={{ borderColor: BORDER_STRONG, color: FG }}
            >
              {t('inventory.addFilter', 'Add filter')}
              <Plus className="h-3 w-3" />
            </button>
          </div>

          <div
            className="grid items-center px-5 py-2 text-[12px] font-medium"
            style={{
              borderTop: `1px solid ${BORDER}`,
              borderBottom: `1px solid ${BORDER}`,
              color: MUTED_FG,
              gridTemplateColumns: '32px 40px 1fr 110px',
            }}
          >
            <div></div>
            <div></div>
            <div>{t('inventory.product', 'Product')}</div>
            <div className="text-right">{t('inventory.sku', 'SKU')}</div>
          </div>

          <div className="flex-1 overflow-y-auto bg-white">
            {filteredGroups.length === 0 ? (
              <div className="px-5 py-12 text-center text-[13px]" style={{ color: MUTED_FG }}>
                {t('inventory.noProductsMatch', 'No products match your search')}
              </div>
            ) : (
              filteredGroups.map((g) => {
                const state = productState(g)
                return (
                  <div key={g.productName}>
                    <button
                      type="button"
                      onClick={() => toggleProduct(g)}
                      className="grid w-full items-center px-5 py-2.5 text-left transition hover:bg-[color:hsl(210_40%_96%/0.7)]"
                      style={{
                        gridTemplateColumns: '32px 40px 1fr 110px',
                        borderBottom: `1px solid ${BORDER}`,
                      }}
                    >
                      <CheckBox state={state} />
                      <div>
                        <div
                          className="grid h-8 w-8 place-items-center rounded-md"
                          style={{
                            background: MUTED,
                            border: `1px solid ${BORDER}`,
                            color: MUTED_FG,
                          }}
                        >
                          <Package className="h-4 w-4" strokeWidth={1.8} />
                        </div>
                      </div>
                      <div className="text-[13px] font-medium" style={{ color: FG }}>
                        {g.productName}
                        {g.variants.length > 1 && (
                          <span
                            className="ml-1.5 text-[12.5px] font-normal"
                            style={{ color: MUTED_FG }}
                          >
                            · {g.variants.length} variants
                          </span>
                        )}
                      </div>
                      <div
                        className="text-right text-[12.5px]"
                        style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                      ></div>
                    </button>
                    {g.variants.length > 1 &&
                      g.variants.map((v) => {
                        const isSel = selected.has(v.id)
                        return (
                          <button
                            type="button"
                            key={v.id}
                            onClick={(e) => {
                              e.stopPropagation()
                              toggleVariant(v.id)
                            }}
                            className="grid w-full items-center px-5 py-2 text-left transition hover:bg-[color:hsl(210_40%_96%/0.7)]"
                            style={{
                              gridTemplateColumns: '32px 40px 1fr 110px',
                              borderBottom: `1px solid ${BORDER}`,
                              background: isSel ? 'hsl(221 83% 53% / 0.05)' : undefined,
                            }}
                          >
                            <div></div>
                            <div className="flex items-center justify-center">
                              <CheckBox state={isSel ? 'all' : 'none'} small />
                            </div>
                            <div className="text-[13px]" style={{ color: FG }}>
                              <span style={{ color: MUTED_FG }}>{g.productName} · </span>
                              {v.variantName}
                            </div>
                            <div
                              className="text-right text-[12.5px]"
                              style={{ color: MUTED_FG, fontVariantNumeric: 'tabular-nums' }}
                            >
                              {v.sku || '—'}
                            </div>
                          </button>
                        )
                      })}
                    {g.variants.length === 1 && (
                      <div
                        className="grid items-center px-5 py-1 text-[12px]"
                        style={{
                          gridTemplateColumns: '32px 40px 1fr 110px',
                          color: MUTED_FG,
                        }}
                      >
                        <div></div>
                        <div></div>
                        <div></div>
                        <div
                          className="text-right"
                          style={{ fontVariantNumeric: 'tabular-nums' }}
                        >
                          {g.variants[0]?.sku || '—'}
                        </div>
                      </div>
                    )}
                  </div>
                )
              })
            )}
          </div>

          <div
            className="flex items-center justify-between gap-3 px-5 py-3"
            style={{ borderTop: `1px solid ${BORDER}` }}
          >
            <div
              className="inline-flex h-[30px] items-center rounded-full px-3 text-[12.5px] font-medium"
              style={{
                background: count > 0 ? 'hsl(221 83% 53% / 0.1)' : MUTED,
                color: count > 0 ? PRIMARY_INK : MUTED_FG,
              }}
            >
              {t('inventory.variantsSelected', {
                count,
                defaultValue: '{{count}} variants selected',
              })}
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => onOpenChange(false)}
                className="inline-flex h-9 items-center rounded-md border bg-white px-3.5 text-[13px] font-medium"
                style={{ borderColor: BORDER, color: FG }}
              >
                {t('common.cancel', 'Cancel')}
              </button>
              <button
                type="button"
                onClick={handleAdd}
                disabled={!canAdd}
                className="inline-flex h-9 items-center rounded-md px-3.5 text-[13px] font-medium text-white"
                style={{
                  background: canAdd ? 'hsl(222.2 47.4% 11.2%)' : 'hsl(220 14% 88%)',
                  color: canAdd ? 'white' : MUTED_FG,
                  cursor: canAdd ? 'pointer' : 'not-allowed',
                }}
              >
                {t('common.add', 'Add')}
              </button>
            </div>
          </div>
        </DialogPrimitive.Content>
      </DialogPortal>
    </DialogPrimitive.Root>
  )
}

function CheckBox({
  state,
  small,
}: {
  state: 'all' | 'some' | 'none'
  small?: boolean
}) {
  const size = small ? 16 : 18
  const checked = state === 'all'
  const indet = state === 'some'
  return (
    <span
      className="relative inline-grid place-items-center rounded-[4px]"
      style={{
        width: size,
        height: size,
        border: `1.5px solid ${checked || indet ? PRIMARY : BORDER_STRONG}`,
        background: checked || indet ? PRIMARY : 'white',
        transition: 'background 100ms, border-color 100ms',
      }}
    >
      {checked && (
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth={3.5}
          strokeLinecap="round"
          strokeLinejoin="round"
          style={{ width: size * 0.66, height: size * 0.66, color: 'white' }}
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
      )}
      {indet && (
        <span
          style={{
            width: size * 0.45,
            height: 2,
            background: 'white',
            borderRadius: 1,
          }}
        />
      )}
    </span>
  )
}
