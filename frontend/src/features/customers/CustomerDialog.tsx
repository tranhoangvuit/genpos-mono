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

import { useCreateCustomer, useCustomerGroups, useUpdateCustomer } from './hooks'
import { customerSchema, type CustomerFormData } from './schemas'
import type { CustomerRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: CustomerRow | null
}

export function CustomerDialog({ open, onOpenChange, existing }: Props) {
  const { t } = useTranslation()
  const create = useCreateCustomer()
  const update = useUpdateCustomer()
  const { data: groups } = useCustomerGroups()

  const form = useForm<CustomerFormData>({
    resolver: standardSchemaResolver(customerSchema(t)),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      notes: '',
      groupIds: [],
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing?.name ?? '',
        email: existing?.email ?? '',
        phone: existing?.phone ?? '',
        notes: existing?.notes ?? '',
        groupIds: existing?.groupIds ?? [],
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
      customer: {
        name: values.name,
        email: values.email,
        phone: values.phone,
        notes: values.notes,
        groupIds: values.groupIds,
      },
    }
    if (existing) {
      await update.mutateAsync({ id: existing.id, ...payload })
    } else {
      await create.mutateAsync(payload)
    }
    onOpenChange(false)
  })

  const selectedGroupIds = form.watch('groupIds')
  const toggleGroup = (id: string) => {
    const curr = new Set(selectedGroupIds)
    if (curr.has(id)) curr.delete(id)
    else curr.add(id)
    form.setValue('groupIds', Array.from(curr), { shouldDirty: true })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {existing ? t('customers.editCustomer') : t('customers.newCustomer')}
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

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="email">{t('customers.email')}</Label>
              <Input id="email" type="email" {...form.register('email')} />
              {form.formState.errors.email && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone">{t('customers.phone')}</Label>
              <Input id="phone" type="tel" {...form.register('phone')} />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">{t('customers.notes')}</Label>
            <Textarea id="notes" rows={3} {...form.register('notes')} />
          </div>

          <div className="space-y-2">
            <Label>{t('customers.groups')}</Label>
            {(groups ?? []).length === 0 ? (
              <p className="text-xs text-[color:var(--color-muted-foreground)]">
                {t('customers.noGroupsHint')}
              </p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {(groups ?? []).map((g) => {
                  const active = selectedGroupIds.includes(g.id)
                  return (
                    <button
                      key={g.id}
                      type="button"
                      onClick={() => toggleGroup(g.id)}
                      className={
                        active
                          ? 'rounded-full border border-[color:var(--color-primary)] bg-[color:var(--color-primary)]/10 px-3 py-1 text-xs font-medium text-[color:var(--color-primary)]'
                          : 'rounded-full border border-[color:var(--color-border)] px-3 py-1 text-xs text-[color:var(--color-muted-foreground)]'
                      }
                    >
                      {g.name}
                    </button>
                  )
                })}
              </div>
            )}
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
