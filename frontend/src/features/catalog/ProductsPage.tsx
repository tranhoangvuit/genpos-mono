import { Link } from '@tanstack/react-router'
import { Plus, Upload } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { DataTable, type DataTableColumn } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'

import { ImportProductDialog } from './ImportProductDialog'
import { useProductList } from './hooks'
import type { ProductListRow } from './types'

export function ProductsPage() {
  const { t } = useTranslation()
  const { data: products, isLoading } = useProductList()
  const [query, setQuery] = useState('')
  const [importOpen, setImportOpen] = useState(false)

  const rows = (products ?? []).filter((r) =>
    query.trim() === ''
      ? true
      : r.name.toLowerCase().includes(query.trim().toLowerCase()) ||
        r.categoryName.toLowerCase().includes(query.trim().toLowerCase()),
  )

  const columns: DataTableColumn<ProductListRow>[] = [
    {
      id: 'name',
      header: t('catalog.name'),
      cell: (r) => (
        <Link
          to="/products/$id"
          params={{ id: r.id }}
          className="font-medium hover:underline"
        >
          {r.name}
        </Link>
      ),
    },
    {
      id: 'category',
      header: t('catalog.category'),
      cell: (r) => r.categoryName || '—',
    },
    {
      id: 'variants',
      header: t('catalog.variants'),
      cell: (r) => r.variantCount,
      headerClassName: 'w-24',
    },
    {
      id: 'price',
      header: t('catalog.price'),
      cell: (r) => formatPrice(r.price),
      headerClassName: 'w-32',
    },
    {
      id: 'status',
      header: t('catalog.status'),
      headerClassName: 'w-24',
      cell: (r) => (
        <span
          className={
            r.isActive
              ? 'inline-flex items-center rounded-md bg-[color:var(--color-success)]/15 px-2 py-0.5 text-xs font-medium text-[color:var(--color-success)]'
              : 'inline-flex items-center rounded-md bg-[color:var(--color-muted)] px-2 py-0.5 text-xs font-medium text-[color:var(--color-muted-foreground)]'
          }
        >
          {r.isActive ? t('catalog.active') : t('catalog.inactive')}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{t('nav.products')}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('catalog.productsSubtitle')}
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setImportOpen(true)}>
            <Upload className="mr-2 h-4 w-4" />
            {t('catalog.importProducts')}
          </Button>
          <Button asChild>
            <Link to="/products/new">
              <Plus className="mr-2 h-4 w-4" />
              {t('catalog.newProduct')}
            </Link>
          </Button>
        </div>
      </div>

      <div className="max-w-sm">
        <Input
          placeholder={t('catalog.searchProducts')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={rows}
        isLoading={isLoading}
        rowKey={(r) => r.id}
        emptyMessage={t('catalog.noProducts')}
      />

      <ImportProductDialog open={importOpen} onOpenChange={setImportOpen} />
    </div>
  )
}

function formatPrice(price: string): string {
  const n = Number(price)
  if (Number.isNaN(n)) return price
  return n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 4 })
}
