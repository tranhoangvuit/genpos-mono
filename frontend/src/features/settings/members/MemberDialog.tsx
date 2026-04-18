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

import { useCreateMember, useRoleOptions, useUpdateMember } from './hooks'
import {
  createMemberSchema,
  updateMemberSchema,
  type CreateMemberFormData,
  type UpdateMemberFormData,
} from './schemas'
import type { MemberRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: MemberRow | null
}

export function MemberDialog({ open, onOpenChange, existing }: Props) {
  if (existing) {
    return <EditMemberDialog open={open} onOpenChange={onOpenChange} existing={existing} />
  }
  return <NewMemberDialog open={open} onOpenChange={onOpenChange} />
}

function NewMemberDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const { data: roles } = useRoleOptions()
  const create = useCreateMember()

  const form = useForm<CreateMemberFormData>({
    resolver: standardSchemaResolver(createMemberSchema(t)),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      roleId: '',
      password: '',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({ name: '', email: '', phone: '', roleId: '', password: '' })
      create.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const errorMessage = create.error ? ConnectError.from(create.error).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    await create.mutateAsync({ member: values })
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('members.newMember')}</DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('members.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="email">{t('members.email')}</Label>
              <Input id="email" type="email" {...form.register('email')} />
              {form.formState.errors.email && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone">{t('members.phone')}</Label>
              <Input id="phone" {...form.register('phone')} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="roleId">{t('members.role')}</Label>
              <select
                id="roleId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('roleId')}
              >
                <option value="">{t('members.selectRole')}</option>
                {(roles ?? []).map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name}
                  </option>
                ))}
              </select>
              {form.formState.errors.roleId && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.roleId.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">{t('members.password')}</Label>
              <Input id="password" type="password" {...form.register('password')} />
              {form.formState.errors.password && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.password.message}
                </p>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={create.isPending}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={create.isPending}>
              {create.isPending ? t('common.saving') : t('common.create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function EditMemberDialog({
  open,
  onOpenChange,
  existing,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: MemberRow
}) {
  const { t } = useTranslation()
  const { data: roles } = useRoleOptions()
  const update = useUpdateMember()

  const form = useForm<UpdateMemberFormData>({
    resolver: standardSchemaResolver(updateMemberSchema(t)),
    defaultValues: {
      name: '',
      phone: '',
      roleId: '',
      status: 'active',
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing.name,
        phone: existing.phone,
        roleId: existing.roleId,
        status: (existing.status as UpdateMemberFormData['status']) ?? 'active',
      })
      update.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, existing])

  const errorMessage = update.error ? ConnectError.from(update.error).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    await update.mutateAsync({ id: existing.id, member: values })
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('members.editMember')}</DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label>{t('members.email')}</Label>
            <Input value={existing.email} disabled readOnly />
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">{t('members.name')}</Label>
            <Input id="name" {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone">{t('members.phone')}</Label>
            <Input id="phone" {...form.register('phone')} />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="roleId">{t('members.role')}</Label>
              <select
                id="roleId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('roleId')}
              >
                {(roles ?? []).map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="status">{t('members.status')}</Label>
              <select
                id="status"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('status')}
              >
                <option value="active">{t('members.statusActive')}</option>
                <option value="inactive">{t('members.statusInactive')}</option>
                <option value="suspended">{t('members.statusSuspended')}</option>
              </select>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={update.isPending}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={update.isPending}>
              {update.isPending ? t('common.saving') : t('common.save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
