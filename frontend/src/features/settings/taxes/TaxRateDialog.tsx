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

import { useCreateTaxRate, useUpdateTaxRate } from './hooks'
import {
  fractionToPercent,
  percentToFraction,
  taxRateSchema,
  type TaxRateFormData,
} from './schemas'
import type { TaxRateRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: TaxRateRow | null
}

export function TaxRateDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreateTaxRate()
  const update = useUpdateTaxRate()

  const form = useForm<TaxRateFormData>({
    resolver: standardSchemaResolver(taxRateSchema(t)),
    defaultValues: {
      name: '',
      ratePercent: '',
      isInclusive: false,
      isDefault: false,
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        ratePercent: existing ? fractionToPercent(existing.rate) : '',
        isInclusive: existing?.isInclusive ?? false,
        isDefault: existing?.isDefault ?? false,
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
      rate: {
        name: values.name,
        rate: percentToFraction(values.ratePercent),
        isInclusive: values.isInclusive,
        isDefault: values.isDefault,
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
            {existing ? t('taxes.editRate') : t('taxes.newRate')}
          </DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('taxes.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="ratePercent">{t('taxes.ratePercent')}</Label>
            <div className="flex items-center gap-2">
              <Input
                id="ratePercent"
                type="number"
                step="0.01"
                min={0}
                max={99.99}
                {...form.register('ratePercent')}
              />
              <span className="text-sm text-[color:var(--color-muted-foreground)]">%</span>
            </div>
            {form.formState.errors.ratePercent && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.ratePercent.message}
              </p>
            )}
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" {...form.register('isInclusive')} />
            {t('taxes.isInclusive')}
          </label>

          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" {...form.register('isDefault')} />
            {t('taxes.isDefault')}
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
