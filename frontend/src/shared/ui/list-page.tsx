import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Filter,
  Plus,
  Search,
} from 'lucide-react'
import { useEffect, useRef, useState, type ReactNode } from 'react'

export const LP_BORDER = 'hsl(214.3 31.8% 91.4%)'
export const LP_MUTED = 'hsl(210 40% 96.1%)'
export const LP_MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
export const LP_FG = 'hsl(222.2 84% 4.9%)'
export const LP_BLUE = 'hsl(221 83% 53%)'
export const LP_BLUE_INK = 'hsl(224 76% 48%)'

export const LP_SWATCHES = [
  '#6B4A2B', '#3A5A40', '#C89B5B', '#4A3527', '#606C38', '#283618',
  '#8A6A4A', '#5C3A1E', '#6A8E4E',
]

export function ListPageShell({ children }: { children: ReactNode }) {
  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      {children}
    </div>
  )
}

export function ListHeader({
  icon,
  title,
  count,
  actions,
}: {
  icon?: ReactNode
  title: ReactNode
  count?: number
  actions?: ReactNode
}) {
  return (
    <div className="mb-[18px] flex items-center justify-between">
      <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
        {icon}
        {title}
        {count !== undefined && (
          <span
            className="rounded-md px-2 py-0.5 text-[13px] font-medium"
            style={{ background: LP_MUTED, color: LP_MUTED_FG }}
          >
            {count}
          </span>
        )}
      </h1>
      {actions && <div className="flex gap-2">{actions}</div>}
    </div>
  )
}

export function PrimaryBtn({
  children,
  onClick,
  icon,
}: {
  children: ReactNode
  onClick?: () => void
  icon?: ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-9 items-center gap-1.5 rounded-md px-3.5 text-[13px] font-medium text-white transition"
      style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
      }}
    >
      {icon ?? <Plus className="h-3.5 w-3.5" strokeWidth={2.2} />}
      {children}
    </button>
  )
}

export function MoreBtn({
  children,
  icon,
  onClick,
}: {
  children: ReactNode
  icon?: ReactNode
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-9 cursor-pointer items-center gap-1.5 rounded-md border bg-white px-3.5 text-[13px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
      style={{ borderColor: LP_BORDER }}
    >
      {icon}
      {children}
    </button>
  )
}

export function IconBtn({
  children,
  onClick,
  label,
}: {
  children: ReactNode
  onClick: () => void
  label: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      className="grid h-7 w-7 place-items-center rounded-md transition hover:bg-[hsl(210_40%_96%)]"
      style={{ color: LP_MUTED_FG }}
    >
      {children}
    </button>
  )
}

export function DeleteBanner({ message }: { message: string | null }) {
  if (!message) return null
  return (
    <div
      role="alert"
      className="mb-3 rounded-md border px-3 py-2 text-sm"
      style={{
        borderColor: 'hsl(0 84% 60% / 0.3)',
        background: 'hsl(0 84% 60% / 0.1)',
        color: 'hsl(0 84% 40%)',
      }}
    >
      {message}
    </div>
  )
}

export function ListSection({ children }: { children: ReactNode }) {
  return (
    <section
      className="overflow-hidden rounded-lg border bg-white"
      style={{ borderColor: LP_BORDER }}
    >
      {children}
    </section>
  )
}

export type Tab = { key: string; label: string; count?: number }

export function ListToolbar<K extends string>({
  tabs,
  activeTab,
  onTabChange,
  filterOpen,
  onToggleFilter,
  sortKey,
  sortDir,
  sortItems,
  onSortKey,
  onSortDir,
}: {
  tabs: Tab[]
  activeTab: string
  onTabChange: (k: string) => void
  filterOpen: boolean
  onToggleFilter: () => void
  sortKey: K
  sortDir: 'asc' | 'desc'
  sortItems: Array<[K, string]>
  onSortKey: (k: K) => void
  onSortDir: (d: 'asc' | 'desc') => void
}) {
  const [sortOpen, setSortOpen] = useState(false)
  const sortRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (!sortRef.current?.contains(e.target as Node)) setSortOpen(false)
    }
    if (sortOpen) document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [sortOpen])

  return (
    <div
      className="flex items-center justify-between border-b px-2.5 pt-1.5"
      style={{ borderColor: LP_BORDER }}
    >
      <div className="flex items-center gap-0.5">
        {tabs.map((tb) => {
          const active = activeTab === tb.key
          return (
            <button
              key={tb.key}
              type="button"
              onClick={() => onTabChange(tb.key)}
              className="-mb-px inline-flex cursor-pointer items-center gap-1.5 rounded-t-md px-3 pb-2.5 pt-2 text-[13px] transition"
              style={{
                color: active ? LP_FG : LP_MUTED_FG,
                fontWeight: active ? 600 : 500,
                borderBottom: `2px solid ${active ? LP_FG : 'transparent'}`,
              }}
            >
              {tb.label}
              {tb.count !== undefined && (
                <span className="text-[11.5px]" style={{ color: LP_MUTED_FG }}>
                  {tb.count}
                </span>
              )}
            </button>
          )
        })}
      </div>
      <div className="flex items-center gap-1 pb-1.5">
        <ToolBtn active={filterOpen} onClick={onToggleFilter} aria-label="Search">
          <Search className="h-[15px] w-[15px]" strokeWidth={2} />
        </ToolBtn>
        <ToolBtn active={filterOpen} onClick={onToggleFilter} aria-label="Filter">
          <Filter className="h-[15px] w-[15px]" strokeWidth={2} />
        </ToolBtn>
        <div className="relative" ref={sortRef}>
          <ToolBtn active={sortOpen} onClick={() => setSortOpen((v) => !v)} aria-label="Sort">
            <ArrowDownUp className="h-[15px] w-[15px]" strokeWidth={2} />
          </ToolBtn>
          {sortOpen && (
            <SortPopover
              sortKey={sortKey}
              sortDir={sortDir}
              items={sortItems}
              onKey={onSortKey}
              onDir={onSortDir}
            />
          )}
        </div>
      </div>
    </div>
  )
}

export function FilterBar({
  placeholder,
  query,
  onQuery,
  onCancel,
  chips,
}: {
  placeholder: string
  query: string
  onQuery: (q: string) => void
  onCancel: () => void
  chips?: string[]
}) {
  return (
    <div className="border-b bg-white px-3 pb-3 pt-2.5" style={{ borderColor: LP_BORDER }}>
      <div className="flex items-center gap-2.5">
        <div
          className="flex h-9 flex-1 items-center gap-2 rounded-md border bg-white px-3"
          style={{
            borderColor: LP_BLUE,
            boxShadow: `0 0 0 3px hsl(221 83% 53% / 0.15)`,
          }}
        >
          <Search className="h-[15px] w-[15px]" strokeWidth={2} style={{ color: LP_MUTED_FG }} />
          <input
            autoFocus
            placeholder={placeholder}
            value={query}
            onChange={(e) => onQuery(e.target.value)}
            className="flex-1 border-none bg-transparent text-[13px] outline-none"
          />
        </div>
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md px-2.5 py-2 text-[13px] font-medium hover:bg-[hsl(210_40%_96%)]"
        >
          Cancel
        </button>
        <button
          type="button"
          className="cursor-default rounded-md px-2.5 py-2 text-[13px] font-medium"
          style={{ color: LP_MUTED_FG }}
        >
          Save as
        </button>
      </div>
      {chips && chips.length > 0 && (
        <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
          {chips.map((c) => (
            <FilterChip key={c}>{c}</FilterChip>
          ))}
          <button
            type="button"
            className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
            style={{ borderColor: LP_BORDER, color: LP_MUTED_FG }}
          >
            <Plus className="h-3 w-3" strokeWidth={2.2} />
            Add filter
          </button>
        </div>
      )}
    </div>
  )
}

export function ListPagination({
  pageStart,
  pageEnd,
  total,
  page,
  onPrev,
  onNext,
}: {
  pageStart: number
  pageEnd: number
  total: number
  page: number
  onPrev: () => void
  onNext: () => void
}) {
  return (
    <div
      className="flex items-center gap-2.5 border-t px-4 py-3 text-[12.5px]"
      style={{
        borderColor: LP_BORDER,
        background: 'hsl(210 40% 96.1% / 0.3)',
        color: LP_MUTED_FG,
      }}
    >
      <PageNavBtn disabled={page === 0} onClick={onPrev}>
        <ChevronLeft className="h-3.5 w-3.5" strokeWidth={2} />
      </PageNavBtn>
      <span>
        {total === 0 ? '0' : `${pageStart + 1} – ${pageEnd}`} of {total}
      </span>
      <div className="flex-1" />
      <PageNavBtn disabled={pageEnd >= total} onClick={onNext}>
        Next
        <ChevronRight className="h-3.5 w-3.5" strokeWidth={2} />
      </PageNavBtn>
    </div>
  )
}

export function ToolBtn({
  children,
  active,
  onClick,
  ...rest
}: {
  children: ReactNode
  active?: boolean
  onClick?: () => void
} & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border bg-white transition hover:bg-[hsl(210_40%_96%)]"
      style={{
        borderColor: active ? `hsl(221 83% 53% / 0.3)` : LP_BORDER,
        background: active ? `hsl(221 83% 53% / 0.1)` : 'white',
        color: active ? LP_BLUE_INK : LP_MUTED_FG,
      }}
      {...rest}
    >
      {children}
    </button>
  )
}

export function SortPopover<K extends string>({
  sortKey,
  sortDir,
  items,
  onKey,
  onDir,
}: {
  sortKey: K
  sortDir: 'asc' | 'desc'
  items: Array<[K, string]>
  onKey: (k: K) => void
  onDir: (d: 'asc' | 'desc') => void
}) {
  return (
    <div
      className="absolute right-0 top-[calc(100%+6px)] z-20 min-w-[240px] rounded-lg border bg-white p-2"
      style={{
        borderColor: LP_BORDER,
        boxShadow:
          '0 16px 40px -12px hsl(222 47% 11% / 0.18), 0 4px 12px hsl(222 47% 11% / 0.06)',
      }}
    >
      <div
        className="px-2 pb-1 pt-1.5 text-[11.5px] font-semibold uppercase tracking-[0.06em]"
        style={{ color: LP_MUTED_FG }}
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
            style={{ color: LP_FG }}
          >
            <span
              className="grid h-4 w-4 flex-shrink-0 place-items-center rounded-full border-[1.5px]"
              style={{ borderColor: on ? LP_BLUE : LP_BORDER }}
            >
              {on && <span className="h-2 w-2 rounded-full" style={{ background: LP_BLUE }} />}
            </span>
            {label}
          </button>
        )
      })}
      <div className="my-1.5 h-px" style={{ background: LP_BORDER }} />
      {(['asc', 'desc'] as const).map((d) => {
        const on = sortDir === d
        const Icon = d === 'asc' ? ArrowUp : ArrowDown
        return (
          <button
            key={d}
            type="button"
            onClick={() => onDir(d)}
            className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-[13px] transition hover:bg-[hsl(210_40%_96%)]"
            style={{
              color: on ? LP_BLUE_INK : LP_FG,
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

export function FilterChip({ children }: { children: ReactNode }) {
  return (
    <button
      type="button"
      className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)]"
      style={{ background: LP_MUTED, borderColor: LP_BORDER, color: LP_FG }}
    >
      {children}
      <ChevronDown className="h-3 w-3" style={{ color: LP_MUTED_FG }} />
    </button>
  )
}

export function Th({
  children,
  align = 'left',
}: {
  children?: ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <th
      className="whitespace-nowrap px-3 py-2.5 text-[12px] font-medium"
      style={{
        textAlign: align,
        color: LP_MUTED_FG,
        background: 'hsl(210 40% 96.1% / 0.4)',
        borderBottom: `1px solid ${LP_BORDER}`,
      }}
    >
      {children}
    </th>
  )
}

export function Td({
  children,
  align = 'left',
}: {
  children?: ReactNode
  align?: 'left' | 'right'
}) {
  return (
    <td
      className="px-3 py-3 text-[13px] align-middle"
      style={{
        textAlign: align,
        color: LP_FG,
        borderBottom: `1px solid ${LP_BORDER}`,
      }}
    >
      {children}
    </td>
  )
}

export function Check({ checked, onClick }: { checked: boolean; onClick: () => void }) {
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
        borderColor: checked ? LP_BLUE : LP_BORDER,
        background: checked ? LP_BLUE : 'white',
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

export function PageNavBtn({
  children,
  disabled,
  onClick,
}: {
  children: ReactNode
  disabled?: boolean
  onClick?: () => void
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="inline-flex cursor-pointer items-center gap-1.5 rounded-md border bg-white px-2.5 py-1 text-[12.5px] font-medium transition hover:bg-[hsl(210_40%_96%)] disabled:cursor-not-allowed disabled:opacity-50"
      style={{ borderColor: LP_BORDER, color: LP_FG }}
    >
      {children}
    </button>
  )
}

export function AllTab({ count, label = 'All' }: { count: number; label?: string }) {
  return (
    <div
      className="flex items-center justify-between border-b px-2.5 pt-1.5"
      style={{ borderColor: LP_BORDER }}
    >
      <div className="flex items-center gap-0.5">
        <button
          type="button"
          className="-mb-px inline-flex cursor-default items-center gap-1.5 rounded-t-md px-3 pb-2.5 pt-2 text-[13px] font-semibold"
          style={{ color: LP_FG, borderBottom: `2px solid ${LP_FG}` }}
        >
          {label}
          <span className="text-[11.5px]" style={{ color: LP_MUTED_FG }}>
            {count}
          </span>
        </button>
      </div>
      <div className="pb-1.5" />
    </div>
  )
}

export function Avatar({
  name,
  index,
  color,
}: {
  name: string
  index: number
  color?: string
}) {
  const bg = color || LP_SWATCHES[index % LP_SWATCHES.length]
  const initial = (name || '?').trim().charAt(0).toUpperCase()
  return (
    <div
      className="grid h-8 w-8 place-items-center overflow-hidden rounded-md text-white"
      style={{ background: bg }}
    >
      <span className="text-[13px] font-semibold">{initial}</span>
    </div>
  )
}
