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

import { useCreateSupplier, useUpdateSupplier } from './hooks'
import { supplierSchema, type SupplierFormData } from './schemas'
import type { SupplierRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: SupplierRow | null
}

export function SupplierDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreateSupplier()
  const update = useUpdateSupplier()

  const form = useForm<SupplierFormData>({
    resolver: standardSchemaResolver(supplierSchema(t)),
    defaultValues: {
      name: '',
      contactName: '',
      email: '',
      phone: '',
      address: '',
      notes: '',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        contactName: existing?.contact_name ?? '',
        email: existing?.email ?? '',
        phone: existing?.phone ?? '',
        address: existing?.address ?? '',
        notes: existing?.notes ?? '',
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
      supplier: {
        name: values.name,
        contactName: values.contactName,
        email: values.email,
        phone: values.phone,
        address: values.address,
        notes: values.notes,
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
            {existing ? t('inventory.editSupplier') : t('inventory.newSupplier')}
          </DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('inventory.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="contactName">{t('inventory.contactName')}</Label>
              <Input id="contactName" {...form.register('contactName')} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">{t('inventory.email')}</Label>
              <Input id="email" type="email" {...form.register('email')} />
              {form.formState.errors.email && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="phone">{t('inventory.phone')}</Label>
              <Input id="phone" type="tel" {...form.register('phone')} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="address">{t('inventory.address')}</Label>
              <Input id="address" {...form.register('address')} />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">{t('inventory.notes')}</Label>
            <Textarea id="notes" rows={2} {...form.register('notes')} />
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
