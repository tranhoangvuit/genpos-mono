import type { ReactNode } from 'react'

import { cn } from '@/shared/lib/cn'

export type DataTableColumn<T> = {
  id: string
  header: ReactNode
  cell: (row: T) => ReactNode
  className?: string
  headerClassName?: string
}

type Props<T> = {
  columns: DataTableColumn<T>[]
  data: T[]
  isLoading?: boolean
  emptyMessage?: ReactNode
  rowKey: (row: T) => string
  onRowClick?: (row: T) => void
  className?: string
}

// DataTable is a simple declarative table built on shadcn/Tailwind tokens.
// It is the shared primitive for list views (products, categories, CSV preview).
export function DataTable<T>({
  columns,
  data,
  isLoading,
  emptyMessage = 'No results',
  rowKey,
  onRowClick,
  className,
}: Props<T>) {
  return (
    <div
      className={cn(
        'overflow-hidden rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)]',
        className,
      )}
    >
      <table className="w-full text-left text-sm">
        <thead className="bg-[color:var(--color-muted)]/40 text-[color:var(--color-muted-foreground)]">
          <tr>
            {columns.map((c) => (
              <th
                key={c.id}
                scope="col"
                className={cn(
                  'px-4 py-3 font-medium text-xs uppercase tracking-wide',
                  c.headerClassName,
                )}
              >
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-[color:var(--color-border)]">
          {isLoading ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-10 text-center text-[color:var(--color-muted-foreground)]"
              >
                Loading...
              </td>
            </tr>
          ) : data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-10 text-center text-[color:var(--color-muted-foreground)]"
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row) => (
              <tr
                key={rowKey(row)}
                className={cn(
                  onRowClick &&
                    'cursor-pointer transition-colors hover:bg-[color:var(--color-accent)]/40',
                )}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
              >
                {columns.map((c) => (
                  <td
                    key={c.id}
                    className={cn('px-4 py-3 align-middle', c.className)}
                  >
                    {c.cell(row)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
