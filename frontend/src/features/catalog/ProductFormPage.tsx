import { useNavigate } from '@tanstack/react-router'
import { ConnectError } from '@connectrpc/connect'
import { Plus, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

import { OptionEditor } from './OptionEditor'
import {
  useCategories,
  useCreateProduct,
  useDeleteProduct,
  useGetProduct,
  useUpdateProduct,
} from './hooks'
import { emptyVariant, generateVariants } from './variants'
import type {
  CategoryRow,
  OptionFormValue,
  ProductFormValues,
  VariantFormValue,
} from './types'

type Props =
  | { mode: 'create'; productId?: undefined }
  | { mode: 'edit'; productId: string }

const defaultValues = (): ProductFormValues => ({
  name: '',
  description: '',
  categoryId: '',
  isActive: true,
  sortOrder: 0,
  hasVariants: false,
  options: [],
  variants: [emptyVariant()],
  images: [],
})

export function ProductFormPage(props: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')
  const toProducts = () =>
    navigate({ to: '/$subdomain/products', params: { subdomain } })

  const { data: categories } = useCategories()
  const createMut = useCreateProduct()
  const updateMut = useUpdateProduct()
  const deleteMut = useDeleteProduct()
  const getMut = useGetProduct()

  const [values, setValues] = useState<ProductFormValues>(defaultValues())
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  // Load product in edit mode
  useEffect(() => {
    if (props.mode !== 'edit') return
    let cancelled = false
    void getMut.mutateAsync(props.productId).then((res) => {
      if (cancelled || !res.product) return
      const p = res.product
      const options: OptionFormValue[] = p.options.map((o) => ({
        name: o.name,
        values: o.values.map((v) => v.value),
      }))
      const valueIdToLabel = new Map<string, string>()
      for (const o of p.options) {
        for (const v of o.values) valueIdToLabel.set(v.id, v.value)
      }
      const variants: VariantFormValue[] = p.variants.map((v, i) => ({
        name: v.name,
        sku: v.sku,
        barcode: v.barcode,
        price: v.price,
        costPrice: v.costPrice,
        trackStock: v.trackStock,
        isActive: v.isActive,
        sortOrder: v.sortOrder || i,
        optionValues: v.optionValueIds.map((id) => valueIdToLabel.get(id) ?? ''),
      }))
      setValues({
        name: p.name,
        description: p.description,
        categoryId: p.categoryId,
        isActive: p.isActive,
        sortOrder: p.sortOrder,
        hasVariants: options.length > 0,
        options,
        variants: variants.length > 0 ? variants : [emptyVariant()],
        images: p.images.map((i) => ({ url: i.url, sortOrder: i.sortOrder })),
      })
    })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.mode, props.mode === 'edit' ? props.productId : undefined])

  const onRegenerateVariants = () => {
    setValues((v) => ({
      ...v,
      variants: generateVariants(v.options, v.variants),
    }))
  }

  const setVariantField = <K extends keyof VariantFormValue>(
    idx: number,
    key: K,
    val: VariantFormValue[K],
  ) => {
    setValues((v) => ({
      ...v,
      variants: v.variants.map((row, i) => (i === idx ? { ...row, [key]: val } : row)),
    }))
  }

  const toggleHasVariants = (on: boolean) => {
    setValues((v) => ({
      ...v,
      hasVariants: on,
      options: on ? (v.options.length > 0 ? v.options : [{ name: '', values: [] }]) : [],
      variants: on ? v.variants : [emptyVariant()],
    }))
  }

  const mutationPending = createMut.isPending || updateMut.isPending || deleteMut.isPending
  const title =
    props.mode === 'create' ? t('catalog.newProduct') : values.name || t('catalog.editProduct')

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrorMessage(null)
    try {
      const req = {
        product: {
          name: values.name.trim(),
          description: values.description,
          categoryId: values.categoryId,
          isActive: values.isActive,
          sortOrder: values.sortOrder,
          options: values.hasVariants
            ? values.options.map((o) => ({ name: o.name, values: o.values }))
            : [],
          variants: (values.hasVariants ? values.variants : values.variants.slice(0, 1)).map(
            (v, i) => ({
              name: v.name || 'Default',
              sku: v.sku,
              barcode: v.barcode,
              price: v.price || '0',
              costPrice: v.costPrice || '0',
              trackStock: v.trackStock,
              isActive: v.isActive,
              sortOrder: i,
              optionValues: values.hasVariants ? v.optionValues : [],
            }),
          ),
          images: values.images.filter((i) => i.url.trim() !== ''),
        },
      }

      if (props.mode === 'create') {
        await createMut.mutateAsync(req)
      } else {
        await updateMut.mutateAsync({ id: props.productId, ...req })
      }
      void toProducts()
    } catch (err) {
      setErrorMessage(ConnectError.from(err).rawMessage)
    }
  }

  const onDelete = async () => {
    if (props.mode !== 'edit') return
    if (!confirm(t('catalog.confirmDelete'))) return
    try {
      await deleteMut.mutateAsync(props.productId)
      void toProducts()
    } catch (err) {
      setErrorMessage(ConnectError.from(err).rawMessage)
    }
  }

  return (
    <form onSubmit={onSubmit} className="mx-auto max-w-5xl space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">{title}</h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {props.mode === 'create' ? t('catalog.newProductSubtitle') : t('catalog.editProductSubtitle')}
          </p>
        </div>
        <div className="flex gap-2">
          {props.mode === 'edit' && (
            <Button
              type="button"
              variant="outline"
              onClick={onDelete}
              disabled={mutationPending}
            >
              {t('common.delete')}
            </Button>
          )}
          <Button
            type="button"
            variant="outline"
            onClick={() => toProducts()}
            disabled={mutationPending}
          >
            {t('common.cancel')}
          </Button>
          <Button type="submit" disabled={mutationPending}>
            {mutationPending ? t('common.saving') : t('common.save')}
          </Button>
        </div>
      </div>

      {errorMessage && (
        <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
          {errorMessage}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main column */}
        <div className="space-y-6 lg:col-span-2">
          <section className="rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">{t('catalog.name')}</Label>
              <Input
                id="name"
                value={values.name}
                onChange={(e) => setValues({ ...values, name: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">{t('catalog.description')}</Label>
              <Textarea
                id="description"
                value={values.description}
                onChange={(e) => setValues({ ...values, description: e.target.value })}
              />
            </div>
          </section>

          <VariantSection
            values={values}
            categories={categories ?? []}
            setValues={setValues}
            setVariantField={setVariantField}
            toggleHasVariants={toggleHasVariants}
            onRegenerate={onRegenerateVariants}
          />

          <section className="rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-semibold">{t('catalog.images')}</h2>
                <p className="text-xs text-[color:var(--color-muted-foreground)]">
                  {t('catalog.imagesSubtitle')}
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  setValues({
                    ...values,
                    images: [...values.images, { url: '', sortOrder: values.images.length }],
                  })
                }
              >
                <Plus className="mr-1 h-4 w-4" />
                {t('catalog.addImage')}
              </Button>
            </div>
            {values.images.length === 0 ? (
              <p className="text-sm text-[color:var(--color-muted-foreground)]">
                {t('catalog.noImages')}
              </p>
            ) : (
              <div className="space-y-2">
                {values.images.map((img, i) => (
                  <div key={i} className="flex gap-2">
                    <Input
                      placeholder="https://..."
                      value={img.url}
                      onChange={(e) =>
                        setValues({
                          ...values,
                          images: values.images.map((row, idx) =>
                            idx === i ? { ...row, url: e.target.value } : row,
                          ),
                        })
                      }
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() =>
                        setValues({
                          ...values,
                          images: values.images.filter((_, idx) => idx !== i),
                        })
                      }
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>

        {/* Sidebar */}
        <aside className="space-y-6">
          <section className="rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="categoryId">{t('catalog.category')}</Label>
              <select
                id="categoryId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                value={values.categoryId}
                onChange={(e) => setValues({ ...values, categoryId: e.target.value })}
              >
                <option value="">{t('catalog.noCategory')}</option>
                {(categories ?? []).map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex items-center gap-2">
              <Checkbox
                id="isActive"
                checked={values.isActive}
                onCheckedChange={(c) => setValues({ ...values, isActive: c === true })}
              />
              <Label htmlFor="isActive" className="cursor-pointer">
                {t('catalog.active')}
              </Label>
            </div>

            <div className="space-y-2">
              <Label htmlFor="sortOrder">{t('catalog.sortOrder')}</Label>
              <Input
                id="sortOrder"
                type="number"
                value={values.sortOrder}
                onChange={(e) => setValues({ ...values, sortOrder: Number(e.target.value) })}
              />
            </div>
          </section>
        </aside>
      </div>
    </form>
  )
}

type VariantSectionProps = {
  values: ProductFormValues
  categories: CategoryRow[]
  setValues: (v: ProductFormValues) => void
  setVariantField: <K extends keyof VariantFormValue>(
    idx: number,
    key: K,
    val: VariantFormValue[K],
  ) => void
  toggleHasVariants: (on: boolean) => void
  onRegenerate: () => void
}

function VariantSection({
  values,
  setValues,
  setVariantField,
  toggleHasVariants,
  onRegenerate,
}: VariantSectionProps) {
  const { t } = useTranslation()
  const variantCount = useMemo(() => values.variants.length, [values.variants])

  return (
    <section className="rounded-xl border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-semibold">{t('catalog.variantsSection')}</h2>
          <p className="text-xs text-[color:var(--color-muted-foreground)]">
            {t('catalog.variantsSubtitle')}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox
            id="hasVariants"
            checked={values.hasVariants}
            onCheckedChange={(c) => toggleHasVariants(c === true)}
          />
          <Label htmlFor="hasVariants" className="cursor-pointer">
            {t('catalog.hasVariants')}
          </Label>
        </div>
      </div>

      {!values.hasVariants ? (
        <div className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="price">{t('catalog.price')}</Label>
            <Input
              id="price"
              value={values.variants[0]?.price ?? '0'}
              onChange={(e) => setVariantField(0, 'price', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="costPrice">{t('catalog.costPrice')}</Label>
            <Input
              id="costPrice"
              value={values.variants[0]?.costPrice ?? '0'}
              onChange={(e) => setVariantField(0, 'costPrice', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="sku">{t('catalog.sku')}</Label>
            <Input
              id="sku"
              value={values.variants[0]?.sku ?? ''}
              onChange={(e) => setVariantField(0, 'sku', e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="barcode">{t('catalog.barcode')}</Label>
            <Input
              id="barcode"
              value={values.variants[0]?.barcode ?? ''}
              onChange={(e) => setVariantField(0, 'barcode', e.target.value)}
            />
          </div>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {values.options.map((opt, i) => (
              <OptionEditor
                key={i}
                option={opt}
                index={i}
                onChange={(next) =>
                  setValues({
                    ...values,
                    options: values.options.map((o, idx) => (idx === i ? next : o)),
                  })
                }
                onRemove={() =>
                  setValues({
                    ...values,
                    options: values.options.filter((_, idx) => idx !== i),
                  })
                }
              />
            ))}
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                setValues({
                  ...values,
                  options: [...values.options, { name: '', values: [] }],
                })
              }
            >
              <Plus className="mr-1 h-4 w-4" />
              {t('catalog.addOption')}
            </Button>
            <Button type="button" onClick={onRegenerate}>
              {t('catalog.regenerateVariants')} ({variantCount})
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="text-[color:var(--color-muted-foreground)]">
                <tr>
                  <th className="px-2 py-2 text-left font-medium">{t('catalog.variant')}</th>
                  <th className="px-2 py-2 text-left font-medium">{t('catalog.sku')}</th>
                  <th className="px-2 py-2 text-left font-medium">{t('catalog.price')}</th>
                  <th className="px-2 py-2 text-left font-medium">{t('catalog.costPrice')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[color:var(--color-border)]">
                {values.variants.map((v, i) => (
                  <tr key={i}>
                    <td className="px-2 py-2 font-medium">{v.name || v.optionValues.join(' / ')}</td>
                    <td className="px-2 py-2">
                      <Input
                        value={v.sku}
                        onChange={(e) => setVariantField(i, 'sku', e.target.value)}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Input
                        value={v.price}
                        onChange={(e) => setVariantField(i, 'price', e.target.value)}
                      />
                    </td>
                    <td className="px-2 py-2">
                      <Input
                        value={v.costPrice}
                        onChange={(e) => setVariantField(i, 'costPrice', e.target.value)}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </section>
  )
}
