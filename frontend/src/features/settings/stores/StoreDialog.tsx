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

import { useCreateStore, useUpdateStore } from './hooks'
import { storeSchema, type StoreFormData } from './schemas'
import type { StoreRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: StoreRow | null
}

export function StoreDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreateStore()
  const update = useUpdateStore()

  const form = useForm<StoreFormData>({
    resolver: standardSchemaResolver(storeSchema(t)),
    defaultValues: {
      name: '',
      address: '',
      phone: '',
      email: '',
      timezone: '',
      status: '',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        address: existing?.address ?? '',
        phone: existing?.phone ?? '',
        email: existing?.email ?? '',
        timezone: existing?.timezone ?? '',
        status: (existing?.status as StoreFormData['status']) ?? '',
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
      store: {
        name: values.name,
        address: values.address,
        phone: values.phone,
        email: values.email,
        timezone: values.timezone,
        status: values.status,
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
      <DialogContent className="w-full max-w-[640px] gap-0 overflow-hidden rounded-[14px] p-0 shadow-[0_20px_48px_rgba(16,24,40,0.22),0_2px_6px_rgba(16,24,40,0.08)]">
        <DialogHeader className="flex flex-row items-center justify-between border-b border-[color:var(--color-border)] bg-[color:var(--color-muted)]/70 px-5 py-4">
          <DialogTitle className="text-[15px] font-semibold tracking-[-0.005em]">
            {existing ? t('stores.editStore') : t('stores.newStore')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={onSubmit} noValidate className="flex flex-col">
          <div className="max-h-[calc(100vh-220px)] space-y-4 overflow-y-auto px-5 py-4">
            {errorMessage && (
              <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
                {errorMessage}
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="name">{t('stores.name')}</Label>
              <Input id="name" autoFocus {...form.register('name')} />
              {form.formState.errors.name && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.name.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="address">{t('stores.address')}</Label>
              <Textarea id="address" rows={2} {...form.register('address')} />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="phone">{t('stores.phone')}</Label>
                <Input id="phone" {...form.register('phone')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">{t('stores.email')}</Label>
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
                <Label htmlFor="timezone">{t('stores.timezone')}</Label>
                <Input
                  id="timezone"
                  placeholder="e.g. Asia/Ho_Chi_Minh"
                  {...form.register('timezone')}
                />
              </div>
              {existing && (
                <div className="space-y-2">
                  <Label htmlFor="status">{t('stores.status')}</Label>
                  <select
                    id="status"
                    className="h-[38px] w-full rounded-lg border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-[13.5px]"
                    {...form.register('status')}
                  >
                    <option value="active">{t('stores.statusActive')}</option>
                    <option value="inactive">{t('stores.statusInactive')}</option>
                    <option value="closed">{t('stores.statusClosed')}</option>
                  </select>
                </div>
              )}
            </div>
          </div>

          <DialogFooter className="flex-row items-center justify-end gap-2 border-t border-[color:var(--color-border)] bg-[color:var(--color-background)] px-5 py-3">
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
