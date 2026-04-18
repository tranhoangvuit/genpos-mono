import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { ConnectError } from '@connectrpc/connect'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

import { useCreateCategory, useUpdateCategory } from './hooks'
import { categorySchema, type CategoryFormData } from './schemas'
import type { CategoryRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: CategoryRow | null
  parents: CategoryRow[]
}

export function CategoryDialog({ open, onOpenChange, existing, parents }: Props) {
  const { t } = useTranslation()
  const create = useCreateCategory()
  const update = useUpdateCategory()

  const form = useForm<CategoryFormData>({
    resolver: standardSchemaResolver(categorySchema(t)),
    defaultValues: {
      name: '',
      parentId: '',
      color: '',
      sortOrder: 0,
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        parentId: existing?.parentId ?? '',
        color: existing?.color ?? '',
        sortOrder: existing?.sortOrder ?? 0,
      })
      create.reset()
      update.reset()
    }
  }, [open, existing, form, create, update])

  const submitting = create.isPending || update.isPending
  const serverError = create.error ?? update.error
  const errorMessage = serverError ? ConnectError.from(serverError).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    if (existing) {
      await update.mutateAsync({
        id: existing.id,
        name: values.name,
        parentId: values.parentId,
        color: values.color,
        sortOrder: values.sortOrder,
      })
    } else {
      await create.mutateAsync({
        name: values.name,
        parentId: values.parentId,
        color: values.color,
        sortOrder: values.sortOrder,
      })
    }
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {existing ? t('catalog.editCategory') : t('catalog.newCategory')}
          </DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('catalog.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="parentId">{t('catalog.parentCategory')}</Label>
            <select
              id="parentId"
              className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
              {...form.register('parentId')}
            >
              <option value="">{t('catalog.noParent')}</option>
              {parents
                .filter((p) => p.id !== existing?.id)
                .map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="color">{t('catalog.color')}</Label>
              <Input id="color" placeholder="#A0C4FF" {...form.register('color')} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="sortOrder">{t('catalog.sortOrder')}</Label>
              <Input
                id="sortOrder"
                type="number"
                {...form.register('sortOrder', { valueAsNumber: true })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting
                ? t('common.saving')
                : existing
                  ? t('common.save')
                  : t('common.create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
