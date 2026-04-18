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
import { Textarea } from '@/shared/ui/textarea'

import { useCreateCustomerGroup, useUpdateCustomerGroup } from './hooks'
import { customerGroupSchema, type CustomerGroupFormData } from './schemas'
import type { CustomerGroupRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: CustomerGroupRow | null
}

export function CustomerGroupDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreateCustomerGroup()
  const update = useUpdateCustomerGroup()

  const form = useForm<CustomerGroupFormData>({
    resolver: standardSchemaResolver(customerGroupSchema(t)),
    defaultValues: {
      name: '',
      description: '',
      discountType: '',
      discountValue: '',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        description: existing?.description ?? '',
        discountType: (existing?.discountType as '' | 'percentage' | 'fixed') ?? '',
        discountValue: existing?.discountValue ?? '',
      })
      create.reset()
      update.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, existing])

  const submitting = create.isPending || update.isPending
  const serverError = create.error ?? update.error
  const errorMessage = serverError ? ConnectError.from(serverError).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    const payload = {
      group: {
        name: values.name,
        description: values.description,
        discountType: values.discountType,
        discountValue: values.discountValue,
      },
    }
    if (existing) {
      await update.mutateAsync({ id: existing.id, ...payload })
    } else {
      await create.mutateAsync(payload)
    }
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {existing ? t('customers.editGroup') : t('customers.newGroup')}
          </DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('customers.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">{t('customers.description')}</Label>
            <Textarea id="description" rows={2} {...form.register('description')} />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="discountType">{t('customers.discountType')}</Label>
              <select
                id="discountType"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('discountType')}
              >
                <option value="">{t('customers.noDiscount')}</option>
                <option value="percentage">{t('customers.discountPercentage')}</option>
                <option value="fixed">{t('customers.discountFixed')}</option>
              </select>
              {form.formState.errors.discountType && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.discountType.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="discountValue">{t('customers.discountValue')}</Label>
              <Input
                id="discountValue"
                type="number"
                step="0.01"
                {...form.register('discountValue')}
              />
              {form.formState.errors.discountValue && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.discountValue.message}
                </p>
              )}
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
