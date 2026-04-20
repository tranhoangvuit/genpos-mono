import { ConnectError } from '@connectrpc/connect'
import {
  ArrowDown,
  ArrowDownUp,
  ArrowUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Download,
  Filter,
  Pencil,
  Plus,
  Search,
  Trash2,
  Truck,
} from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'

import { SupplierDialog } from './SupplierDialog'
import { useDeleteSupplier, useSuppliers } from './hooks'
import type { SupplierRow } from './types'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const BLUE = 'hsl(221 83% 53%)'
const BLUE_INK = 'hsl(224 76% 48%)'

const PAGE_SIZE = 15

const SWATCHES = [
  '#6B4A2B', '#3A5A40', '#C89B5B', '#4A3527', '#606C38', '#283618',
  '#8A6A4A', '#5C3A1E', '#6A8E4E',
]

type SortKey = 'name' | 'contact' | 'email' | 'updated'
type SortDir = 'asc' | 'desc'

export function SuppliersPage() {
  const { t } = useTranslation()
  const { data: suppliers, isLoading } = useSuppliers()
  const deleteMut = useDeleteSupplier()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<SupplierRow | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [pendingDelete, setPendingDelete] = useState<SupplierRow | null>(null)
  const [query, setQuery] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [sortOpen, setSortOpen] = useState(false)
  const [sortKey, setSortKey] = useState<SortKey>('name')
  const [sortDir, setSortDir] = useState<SortDir>('asc')
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

  const list = suppliers ?? []

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    let arr = list
    if (q) {
      arr = arr.filter(
        (s) =>
          s.name.toLowerCase().includes(q) ||
          (s.contactName ?? '').toLowerCase().includes(q) ||
          (s.email ?? '').toLowerCase().includes(q) ||
          (s.phone ?? '').toLowerCase().includes(q),
      )
    }
    arr = [...arr].sort((a, b) => {
      let cmp = 0
      switch (sortKey) {
        case 'contact':
          cmp = (a.contactName ?? '').localeCompare(b.contactName ?? '')
          break
        case 'email':
          cmp = (a.email ?? '').localeCompare(b.email ?? '')
          break
        case 'updated': {
          const ax = a.updatedAt ? Number(a.updatedAt.seconds) : 0
          const bx = b.updatedAt ? Number(b.updatedAt.seconds) : 0
          cmp = ax - bx
          break
        }
        default:
          cmp = a.name.localeCompare(b.name)
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return arr
  }, [list, query, sortKey, sortDir])

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

  const confirmDelete = async () => {
    if (!pendingDelete) return
    const row = pendingDelete
    try {
      await deleteMut.mutateAsync(row.id)
      setPendingDelete(null)
    } catch (err) {
      setDeleteError(ConnectError.from(err).rawMessage)
      setPendingDelete(null)
    }
  }

  return (
    <div className="mx-auto -my-6 py-6" style={{ maxWidth: 1600 }}>
      <div className="mb-[18px] flex items-center justify-between">
        <h1 className="m-0 flex items-center gap-2.5 text-[20px] font-semibold tracking-[-0.005em]">
          <Truck className="h-[18px] w-[18px]" strokeWidth={2} />
          {t('nav.suppliers')}
          <span
            className="rounded-md px-2 py-0.5 text-[13px] font-medium"
            style={{ background: MUTED, color: MUTED_FG }}
          >
            {list.length}
          </span>
        </h1>
        <div className="flex gap-2">
          <MoreBtn icon={<Download className="h-3.5 w-3.5" />}>Export</MoreBtn>
          <button
            type="button"
            onClick={() => {
              setEditing(null)
              setDialogOpen(true)
            }}
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
            {t('inventory.newSupplier')}
          </button>
        </div>
      </div>

      {deleteError && (
        <div
          role="alert"
          className="mb-3 rounded-md border px-3 py-2 text-sm"
          style={{
            borderColor: 'hsl(0 84% 60% / 0.3)',
            background: 'hsl(0 84% 60% / 0.1)',
            color: 'hsl(0 84% 40%)',
          }}
        >
          {deleteError}
        </div>
      )}

      <section
        className="overflow-hidden rounded-lg border bg-white"
        style={{ borderColor: BORDER }}
      >
        <div
          className="flex items-center justify-between border-b px-2.5 pt-1.5"
          style={{ borderColor: BORDER }}
        >
          <div className="flex items-center gap-0.5">
            <button
              type="button"
              className="-mb-px inline-flex cursor-default items-center gap-1.5 rounded-t-md px-3 pb-2.5 pt-2 text-[13px] font-semibold"
              style={{
                color: FG,
                borderBottom: `2px solid ${FG}`,
              }}
            >
              All
              <span className="text-[11.5px]" style={{ color: MUTED_FG }}>
                {list.length}
              </span>
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
                  placeholder="Searching all suppliers"
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
            </div>
            <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
              <FilterChip>Tag</FilterChip>
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
          <table className="w-full border-collapse" style={{ minWidth: 960 }}>
            <colgroup>
              <col style={{ width: 44 }} />
              <col style={{ width: 48 }} />
              <col />
              <col style={{ width: 200 }} />
              <col style={{ width: 240 }} />
              <col style={{ width: 160 }} />
              <col style={{ width: 88 }} />
            </colgroup>
            <thead>
              <tr>
                <Th>
                  <Check checked={allSelectedOnPage} onClick={toggleAll} />
                </Th>
                <Th />
                <Th>{t('inventory.name')}</Th>
                <Th>{t('inventory.contactName')}</Th>
                <Th>{t('inventory.email')}</Th>
                <Th>{t('inventory.phone')}</Th>
                <Th />
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
                    {t('inventory.noSuppliers')}
                  </td>
                </tr>
              )}
              {pageRows.map((s, i) => (
                <Row
                  key={s.id}
                  s={s}
                  i={pageStart + i}
                  checked={selected.has(s.id)}
                  onToggle={() => toggle(s.id)}
                  onEdit={() => {
                    setEditing(s)
                    setDialogOpen(true)
                  }}
                  onDelete={() => {
                    setDeleteError(null)
                    setPendingDelete(s)
                  }}
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

      <SupplierDialog open={dialogOpen} onOpenChange={setDialogOpen} existing={editing} />

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
          <p className="text-sm" style={{ color: MUTED_FG }}>
            {pendingDelete
              ? t('inventory.confirmDeleteSupplier', { name: pendingDelete.name })
              : ''}
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
    </div>
  )
}

function Row({
  s,
  i,
  checked,
  onToggle,
  onEdit,
  onDelete,
}: {
  s: SupplierRow
  i: number
  checked: boolean
  onToggle: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  const bg = SWATCHES[i % SWATCHES.length]
  const initial = (s.name || '?').trim().charAt(0).toUpperCase()
  return (
    <tr className="group transition hover:bg-[hsl(210_40%_96%_/_0.3)]">
      <Td>
        <Check checked={checked} onClick={onToggle} />
      </Td>
      <Td>
        <div
          className="grid h-8 w-8 place-items-center overflow-hidden rounded-md text-white"
          style={{ background: bg }}
        >
          <span className="text-[13px] font-semibold">{initial}</span>
        </div>
      </Td>
      <Td>
        <button
          type="button"
          onClick={onEdit}
          className="font-medium hover:underline"
          style={{ color: FG, textDecorationColor: MUTED_FG }}
        >
          {s.name}
        </button>
      </Td>
      <Td>
        <span style={{ color: s.contactName ? FG : MUTED_FG }}>
          {s.contactName || '—'}
        </span>
      </Td>
      <Td>
        <span style={{ color: s.email ? FG : MUTED_FG }}>{s.email || '—'}</span>
      </Td>
      <Td>
        <span className="tabular-nums" style={{ color: s.phone ? FG : MUTED_FG }}>
          {s.phone || '—'}
        </span>
      </Td>
      <Td align="right">
        <div className="flex items-center justify-end gap-1 opacity-0 transition group-hover:opacity-100">
          <IconBtn onClick={onEdit} label="Edit">
            <Pencil className="h-3.5 w-3.5" strokeWidth={2} />
          </IconBtn>
          <IconBtn onClick={onDelete} label="Delete">
            <Trash2 className="h-3.5 w-3.5" strokeWidth={2} />
          </IconBtn>
        </div>
      </Td>
    </tr>
  )
}

function IconBtn({
  children,
  onClick,
  label,
}: {
  children: React.ReactNode
  onClick: () => void
  label: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      className="grid h-7 w-7 place-items-center rounded-md transition hover:bg-[hsl(210_40%_96%)]"
      style={{ color: MUTED_FG }}
    >
      {children}
    </button>
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
    ['name', 'Name'],
    ['contact', 'Contact'],
    ['email', 'Email'],
    ['updated', 'Updated'],
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
