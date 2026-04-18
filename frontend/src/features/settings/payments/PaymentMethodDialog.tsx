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

import { useCreatePaymentMethod, useUpdatePaymentMethod } from './hooks'
import { paymentMethodSchema, type PaymentMethodFormData } from './schemas'
import { PAYMENT_METHOD_TYPES, type PaymentMethodRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: PaymentMethodRow | null
}

export function PaymentMethodDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreatePaymentMethod()
  const update = useUpdatePaymentMethod()

  const form = useForm<PaymentMethodFormData>({
    resolver: standardSchemaResolver(paymentMethodSchema(t)),
    defaultValues: {
      name: '',
      type: 'cash',
      isActive: true,
      sortOrder: 0,
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        type: (existing?.type as PaymentMethodFormData['type']) ?? 'cash',
        isActive: existing?.isActive ?? true,
        sortOrder: existing?.sortOrder ?? 0,
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
      method: {
        name: values.name,
        type: values.type,
        isActive: values.isActive,
        sortOrder: values.sortOrder,
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
            {existing ? t('payments.editMethod') : t('payments.newMethod')}
          </DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('payments.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="type">{t('payments.type')}</Label>
              <select
                id="type"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('type')}
              >
                {PAYMENT_METHOD_TYPES.map((v) => (
                  <option key={v} value={v}>
                    {t(`payments.type_${v}`)}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="sortOrder">{t('payments.sortOrder')}</Label>
              <Input
                id="sortOrder"
                type="number"
                min={0}
                {...form.register('sortOrder', { valueAsNumber: true })}
              />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" {...form.register('isActive')} />
            {t('payments.isActive')}
          </label>

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
