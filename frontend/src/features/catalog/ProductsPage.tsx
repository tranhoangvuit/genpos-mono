import { Link } from '@tanstack/react-router'
import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  Box,
  Calendar,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Download,
  Filter,
  HelpCircle,
  Package,
  Plus,
  Search,
  Upload,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'

import { ImportProductDialog } from './ImportProductDialog'
import { useProductList } from './hooks'
import type { ProductListRow } from './types'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const BLUE_SOFT = 'hsl(214 100% 96.7%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE = 'hsl(142.1 76.2% 36.3%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const IDLE_SOFT = 'hsl(210 40% 96.1%)'
const IDLE_INK = 'hsl(215.3 25% 26.7%)'

type TabKey = 'all' | 'active' | 'draft' | 'archived'
type SortKey = 'title' | 'created' | 'updated' | 'inventory' | 'type' | 'publish' | 'vendor'
type SortDir = 'asc' | 'desc'

const PAGE_SIZE = 15

const ICON_COLORS = [
  'hsl(25 95% 53%)',
  'hsl(142 76% 36%)',
  'hsl(221 83% 53%)',
  'hsl(261 83% 58%)',
  'hsl(346 77% 49%)',
  'hsl(197 92% 45%)',
]

export function ProductsPage() {
  const { t } = useTranslation()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const { data: products, isLoading } = useProductList()
  const [query, setQuery] = useState('')
  const [tab, setTab] = useState<TabKey>('all')
  const [sortKey, setSortKey] = useState<SortKey>('created')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [importOpen, setImportOpen] = useState(false)
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const sortRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
    }
    if (sortOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen])

  const all = products ?? []
  const counts = useMemo(
    () => ({
      all: all.length,
      active: all.filter((p) => p.isActive).length,
      draft: all.filter((p) => !p.isActive).length,
      archived: 0,
    }),
    [all],
  )

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let list = all
    if (tab === 'active') list = list.filter((p) => p.isActive)
    else if (tab === 'draft') list = list.filter((p) => !p.isActive)
    else if (tab === 'archived') list = []
    if (q) {
      list = list.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          p.categoryName.toLowerCase().includes(q),
      )
    }
    list = [...list].sort((a, b) => {
      const cmp = a.name.localeCompare(b.name)
      return sortDir === 'asc' ? cmp : -cmp
    })
    return list
  }, [all, query, tab, sortDir])

  const pageStart = page * PAGE_SIZE
  const pageRows = filtered.slice(pageStart, pageStart + PAGE_SIZE)
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

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <Package className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('nav.products')}
          <span
            className="rounded-md px-2 py-0.5 text-[13px] font-medium"
            style={{ background: MUTED, color: MUTED_FG }}
          >
            {all.length}
          </span>
        </h1>
        <div className="flex gap-2">
          <MoreBtn icon={<Download className="h-3.5 w-3.5" />}>Export</MoreBtn>
          <MoreBtn
            icon={<Upload className="h-3.5 w-3.5" />}
            onClick={() => setImportOpen(true)}
          >
            {t('catalog.importProducts')}
          </MoreBtn>
          <MoreBtn>
            More actions
            <ChevronDown className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
          </MoreBtn>
          <Link
            to="/$subdomain/products/new"
            params={{ subdomain }}
            className="inline-flex h-9 items-center gap-1.5 rounded-md px-3.5 text-[13px] font-medium text-white transition"
            style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
            }}
          >
            <Plus className="h-3.5 w-3.5" strokeWidth={2.2} />
            {t('catalog.newProduct')}
          </Link>
        </div>
      </div>

      <section
        className="mb-4 grid grid-cols-[auto_repeat(3,1fr)] rounded-lg border bg-white p-0.5"
        style={{ borderColor: BORDER }}
      >
        <KpiCell divider={false}>
          <button
            type="button"
            className="inline-flex items-center gap-1.5 rounded-md border bg-white px-2.5 py-1.5 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
            style={{ borderColor: BORDER }}
          >
            <Calendar className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
            30 days
          </button>
        </KpiCell>
        <KpiCell>
          <KpiLabel>Average sell-through rate</KpiLabel>
          <div className="mt-1 flex items-baseline gap-1.5 text-[20px] font-semibold tabular-nums">
            0%
            <span className="text-[14px] font-normal" style={{ color: MUTED_FG }}>
              —
            </span>
          </div>
        </KpiCell>
        <KpiCell>
          <KpiLabel>Products by days of inventory remaining</KpiLabel>
          <div className="mt-1 text-[14px] font-medium" style={{ color: MUTED_FG }}>
            No data
          </div>
        </KpiCell>
        <KpiCell>
          <KpiLabel>ABC product analysis</KpiLabel>
          <div className="mt-1 flex items-baseline gap-1.5 text-[20px] font-semibold tabular-nums">
            <span
              className="underline decoration-2 underline-offset-[3px]"
              style={{ textDecorationColor: BLUE }}
            >
              $0.00
            </span>
            <span
              className="cursor-pointer text-[12px] font-medium"
              style={{ color: BLUE }}
            >
              C
            </span>
          </div>
        </KpiCell>
      </section>

      <section
        className="overflow-hidden rounded-lg border bg-white"
        style={{ borderColor: BORDER }}
      >
        <div
          className="flex items-center justify-between border-b px-2.5 pt-1.5"
          style={{ borderColor: BORDER }}
        >
          <div className="flex items-center gap-0.5">
            {(['all', 'active', 'draft', 'archived'] as TabKey[]).map((k) => (
              <button
                key={k}
                type="button"
                onClick={() => {
                  setTab(k)
                  setPage(0)
                }}
                className="-mb-px inline-flex cursor-pointer items-center gap-1.5 rounded-t-md px-3 pb-2.5 pt-2 text-[13px] transition"
                style={{
                  color: tab === k ? FG : MUTED_FG,
                  fontWeight: tab === k ? 600 : 500,
                  borderBottom: `2px solid ${tab === k ? FG : 'transparent'}`,
                }}
              >
                {k === 'all' ? 'All' : k === 'active' ? t('catalog.active') : k === 'draft' ? 'Draft' : 'Archived'}
                {k !== 'archived' && (
                  <span className="text-[11.5px]" style={{ color: MUTED_FG }}>
                    {counts[k]}
                  </span>
                )}
              </button>
            ))}
            <button
              type="button"
              className="inline-flex cursor-pointer items-center rounded-md px-2 py-1.5 transition hover:bg-[hsl(210_40%_96%)]"
              style={{ color: MUTED_FG }}
              aria-label="Add view"
            >
              <Plus className="h-3.5 w-3.5" strokeWidth={2.2} />
            </button>
          </div>
          <div className="flex items-center gap-1 pb-1.5">
            <ToolBtn
              active={filterOpen}
              onClick={() => setFilterOpen((v) => !v)}
              aria-label="Search"
            >
              <Search className="h-[15px] w-[15px]" strokeWidth={2} />
            </ToolBtn>
            <ToolBtn
              active={filterOpen}
              onClick={() => setFilterOpen((v) => !v)}
              aria-label="Filter"
            >
              <Filter className="h-[15px] w-[15px]" strokeWidth={2} />
            </ToolBtn>
            <div className="relative" ref={sortRef}>
              <ToolBtn
                active={sortOpen}
                onClick={() => setSortOpen((v) => !v)}
                aria-label="Sort"
              >
                <ArrowDownUp className="h-[15px] w-[15px]" strokeWidth={2} />
              </ToolBtn>
              {sortOpen && (
                <SortPopover
                  sortKey={sortKey}
                  sortDir={sortDir}
                  onKey={setSortKey}
                  onDir={setSortDir}
                />
              )}
            </div>
          </div>
        </div>

        {filterOpen && (
          <div
            className="border-b bg-white px-3 pb-3 pt-2.5"
            style={{ borderColor: BORDER }}
          >
            <div className="flex items-center gap-2.5">
              <div
                className="flex h-9 flex-1 items-center gap-2 rounded-md border bg-white px-3"
                style={{
                  borderColor: BLUE,
                  boxShadow: `0 0 0 3px hsl(221 83% 53% / 0.15)`,
                }}
              >
                <Search
                  className="h-[15px] w-[15px]"
                  strokeWidth={2}
                  style={{ color: MUTED_FG }}
                />
                <input
                  autoFocus
                  placeholder="Searching all products"
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value)
                    setPage(0)
                  }}
                  className="flex-1 border-none bg-transparent text-[13px] outline-none"
                />
              </div>
              <button
                type="button"
                onClick={() => {
                  setFilterOpen(false)
                  setQuery('')
                  setPage(0)
                }}
                className="rounded-md px-2.5 py-2 text-[13px] font-medium hover:bg-[hsl(210_40%_96%)]"
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                className="cursor-default rounded-md px-2.5 py-2 text-[13px] font-medium"
                style={{ color: MUTED_FG }}
              >
                Save as
              </button>
              <button
                type="button"
                className="grid h-8 w-8 place-items-center rounded-md hover:bg-[hsl(210_40%_96%)]"
                style={{ color: MUTED_FG }}
                aria-label="Saved views"
              >
                <ArrowDownUp className="h-[15px] w-[15px]" strokeWidth={2} />
              </button>
            </div>
            <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
              <FilterChip>Vendors</FilterChip>
              <FilterChip>Tag</FilterChip>
              <FilterChip>Statuses</FilterChip>
              <button
                type="button"
                className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
                style={{ borderColor: BORDER, color: MUTED_FG }}
              >
                <Plus className="h-3 w-3" strokeWidth={2.2} />
                Add filter
              </button>
            </div>
          </div>
        )}

        <div className="w-full overflow-x-auto">
          <table className="w-full border-collapse" style={{ tableLayout: 'fixed' }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 48 }} />
              <col />
              <col style={{ width: 110 }} />
              <col style={{ width: 180 }} />
              <col style={{ width: 180 }} />
              <col style={{ width: 100 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th />
                <Th>Product</Th>
                <Th>Status</Th>
                <Th>Inventory</Th>
                <Th>Category</Th>
                <Th align="right">Channels</Th>
              </tr>
            </thead>
            <tbody>
              {isLoading && (
                <tr>
                  <td
                    colSpan={7}
                    className="py-10 text-center text-[13px]"
                    style={{ color: MUTED_FG }}
                  >
                    {t('common.loading')}
                  </td>
                </tr>
              )}
              {!isLoading && pageRows.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="py-10 text-center text-[13px]"
                    style={{ color: MUTED_FG }}
                  >
                    {t('catalog.noProducts')}
                  </td>
                </tr>
              )}
              {pageRows.map((p, i) => (
                <Row
                  key={p.id}
                  p={p}
                  i={pageStart + i}
                  subdomain={subdomain}
                  checked={selected.has(p.id)}
                  onToggle={() => toggle(p.id)}
                />
              ))}
            </tbody>
          </table>
        </div>

        <div
          className="flex items-center gap-2.5 border-t px-4 py-3 text-[12.5px]"
          style={{
            borderColor: BORDER,
            background: 'hsl(210 40% 96.1% / 0.3)',
            color: MUTED_FG,
          }}
        >
          <PageNavBtn
            disabled={page === 0}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
          >
            <ChevronLeft className="h-3.5 w-3.5" strokeWidth={2} />
          </PageNavBtn>
          <span>
            {filtered.length === 0 ? '0' : `${pageStart + 1} – ${pageEnd}`} of{' '}
            {filtered.length}
          </span>
          <div className="flex-1" />
          <PageNavBtn
            disabled={pageEnd >= filtered.length}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
            <ChevronRight className="h-3.5 w-3.5" strokeWidth={2} />
          </PageNavBtn>
        </div>
      </section>

      <ImportProductDialog open={importOpen} onOpenChange={setImportOpen} />
    </div>
  )
}

function KpiCell({
  children,
  divider = true,
}: {
  children: React.ReactNode
  divider?: boolean
}) {
  return (
    <div
      className="min-w-0 px-4 py-3.5"
      style={{
        borderLeft: divider ? `1px solid ${BORDER}` : undefined,
      }}
    >
      {children}
    </div>
  )
}

function KpiLabel({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="inline-flex items-center gap-1 text-[12.5px] font-medium"
      style={{ color: MUTED_FG }}
    >
      {children}
      <HelpCircle className="h-3 w-3" strokeWidth={2} />
    </div>
  )
}

function MoreBtn({
  children,
  icon,
  onClick,
}: {
  children: React.ReactNode
  icon?: React.ReactNode
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-9 cursor-pointer items-center gap-1.5 rounded-md border bg-white px-3.5 text-[13px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
      style={{ borderColor: BORDER }}
    >
      {icon}
      {children}
    </button>
  )
}

function ToolBtn({
  children,
  active,
  onClick,
  ...rest
}: {
  children: React.ReactNode
  active?: boolean
  onClick?: () => void
} & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border bg-white transition hover:bg-[hsl(210_40%_96%)]"
      style={{
        borderColor: active ? `hsl(221 83% 53% / 0.3)` : BORDER,
        background: active ? `hsl(221 83% 53% / 0.1)` : 'white',
        color: active ? BLUE_INK : MUTED_FG,
      }}
      {...rest}
    >
      {children}
    </button>
  )
}

function SortPopover({
  sortKey,
  sortDir,
  onKey,
  onDir,
}: {
  sortKey: SortKey
  sortDir: SortDir
  onKey: (k: SortKey) => void
  onDir: (d: SortDir) => void
}) {
  const items: Array<[SortKey, string]> = [
    ['title', 'Product title'],
    ['created', 'Created'],
    ['updated', 'Updated'],
    ['inventory', 'Inventory'],
    ['type', 'Product type'],
    ['publish', 'Publishing error'],
    ['vendor', 'Vendor'],
  ]
  return (
    <div
      className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[240px] rounded-lg border bg-white p-2"
      style={{
        borderColor: BORDER,
        boxShadow:
          '0 16px 40px -12px hsl(222 47% 11% / 0.18), 0 4px 12px hsl(222 47% 11% / 0.06)',
      }}
    >
      <div
        className="px-2 pb-1 pt-1.5 text-[11.5px] font-semibold uppercase tracking-[0.06em]"
        style={{ color: MUTED_FG }}
      >
        Sort by
      </div>
      {items.map(([k, label]) => {
        const on = sortKey === k
        return (
          <button
            key={k}
            type="button"
            onClick={() => onKey(k)}
            className="flex w-full cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-left text-[13px] transition hover:bg-[hsl(210_40%_96%)]"
            style={{ color: FG }}
          >
            <span
              className="grid h-4 w-4 flex-shrink-0 place-items-center rounded-full border-[1.5px]"
              style={{ borderColor: on ? BLUE : BORDER }}
            >
              {on && (
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ background: BLUE }}
                />
              )}
            </span>
            {label}
          </button>
        )
      })}
      <div className="my-1.5 h-px" style={{ background: BORDER }} />
      {(['asc', 'desc'] as SortDir[]).map((d) => {
        const on = sortDir === d
        const Icon = d === 'asc' ? ArrowUp : ArrowDown
        return (
          <button
            key={d}
            type="button"
            onClick={() => onDir(d)}
            className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-[13px] transition hover:bg-[hsl(210_40%_96%)]"
            style={{
              color: on ? BLUE_INK : FG,
              fontWeight: on ? 600 : 400,
            }}
          >
            <Icon className="h-[13px] w-[13px]" strokeWidth={2.2} />
            {d === 'asc' ? 'Oldest first' : 'Newest first'}
          </button>
        )
      })}
    </div>
  )
}

function FilterChip({ children }: { children: React.ReactNode }) {
  return (
    <button
      type="button"
      className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
      style={{ background: MUTED, borderColor: BORDER, color: FG }}
    >
      {children}
      <ChevronDown className="h-3 w-3" style={{ color: MUTED_FG }} />
    </button>
  )
}

function Th({
  children,
  align = 'left',
}: {
  children?: React.ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <th
      className="whitespace-nowrap px-3 py-2.5 text-[12px] font-medium"
      style={{
        textAlign: align,
        color: MUTED_FG,
        background: 'hsl(210 40% 96.1% / 0.4)',
        borderBottom: `1px solid ${BORDER}`,
      }}
    >
      {children}
    </th>
  )
}

function Check({
  checked,
  onClick,
}: {
  checked: boolean
  onClick: () => void
}) {
  return (
    <span
      role="checkbox"
      aria-checked={checked}
      onClick={(e) => {
        e.stopPropagation()
        onClick()
      }}
      className="inline-grid h-4 w-4 cursor-pointer place-items-center rounded border-[1.5px] transition-colors"
      style={{
        borderColor: checked ? BLUE : BORDER,
        background: checked ? BLUE : 'white',
        color: 'white',
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
          className="h-[11px] w-[11px]"
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
      )}
    </span>
  )
}

function Row({
  p,
  i,
  subdomain,
  checked,
  onToggle,
}: {
  p: ProductListRow
  i: number
  subdomain: string
  checked: boolean
  onToggle: () => void
}) {
  const status = p.isActive ? 'active' : 'draft'
  const thumbBg = ICON_COLORS[i % ICON_COLORS.length]
  return (
    <tr className="transition hover:bg-[hsl(210_40%_96%_/_0.3)]">
      <Td>
        <Check checked={checked} onClick={onToggle} />
      </Td>
      <Td>
        <div
          className="grid h-8 w-8 place-items-center overflow-hidden rounded-md text-white"
          style={{ background: thumbBg }}
        >
          <Box className="h-[15px] w-[15px]" strokeWidth={2} />
        </div>
      </Td>
      <Td>
        <Link
          to="/$subdomain/products/$id"
          params={{ subdomain, id: p.id }}
          className="font-medium hover:underline"
          style={{ color: FG, textDecorationColor: MUTED_FG }}
        >
          {p.name}
        </Link>
      </Td>
      <Td>
        <StatusPill status={status} />
      </Td>
      <Td>
        <span
          className="tabular-nums"
          style={{ color: p.variantCount === 0 ? MUTED_FG : FG }}
        >
          {p.variantCount} {p.variantCount === 1 ? 'variant' : 'variants'}
        </span>
      </Td>
      <Td>
        {p.categoryName ? (
          <span
            className="inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-[12px]"
            style={{ background: MUTED, color: FG }}
          >
            {p.categoryName}
          </span>
        ) : (
          <span style={{ color: MUTED_FG }}>—</span>
        )}
      </Td>
      <Td align="right">
        <span
          className="tabular-nums font-medium"
          style={{ color: FG }}
        >
          1
        </span>
      </Td>
    </tr>
  )
}

function Td({
  children,
  align = 'left',
}: {
  children?: React.ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <td
      className="px-3 py-3 text-[13px] align-middle"
      style={{
        textAlign: align,
        color: FG,
        borderBottom: `1px solid ${BORDER}`,
      }}
    >
      {children}
    </td>
  )
}

function StatusPill({ status }: { status: 'active' | 'draft' | 'archived' }) {
  const palette =
    status === 'active'
      ? { bg: DONE_SOFT, color: DONE_INK, dot: DONE }
      : status === 'draft'
        ? { bg: IDLE_SOFT, color: IDLE_INK, dot: MUTED_FG }
        : { bg: MUTED, color: MUTED_FG, dot: MUTED_FG }
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2 py-[2px] text-[11.5px] font-medium"
      style={{ background: palette.bg, color: palette.color }}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: palette.dot }}
      />
      {status === 'active' ? 'Active' : status === 'draft' ? 'Draft' : 'Archived'}
    </span>
  )
}

function PageNavBtn({
  children,
  disabled,
  onClick,
}: {
  children: React.ReactNode
  disabled?: boolean
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-white px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)] disabled:cursor-not-allowed disabled:opacity-50"
      style={{ borderColor: BORDER, color: FG }}
    >
      {children}
    </button>
  )
}
