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

import { useStores } from '@/features/settings/stores/hooks'

import { useCreateStaff, useRoleOptions, useUpdateStaff } from './hooks'
import {
  createStaffSchema,
  updateStaffSchema,
  type CreateStaffFormData,
  type UpdateStaffFormData,
} from './schemas'
import type { StaffRow } from './types'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: StaffRow | null
}

export function StaffDialog({ open, onOpenChange, existing }: Props) {
  if (existing) {
    return <EditStaffDialog open={open} onOpenChange={onOpenChange} existing={existing} />
  }
  return <NewStaffDialog open={open} onOpenChange={onOpenChange} />
}

function NewStaffDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const { data: roles } = useRoleOptions()
  const create = useCreateStaff()

  const form = useForm<CreateStaffFormData>({
    resolver: standardSchemaResolver(createStaffSchema(t)),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      roleId: '',
      password: '',
      allStores: true,
      storeIds: [],
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: '',
        email: '',
        phone: '',
        roleId: '',
        password: '',
        allStores: true,
        storeIds: [],
      })
      create.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const errorMessage = create.error ? ConnectError.from(create.error).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    await create.mutateAsync({
      member: {
        ...values,
        storeIds: values.allStores ? [] : values.storeIds,
      },
    })
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('staffs.newStaff')}</DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label htmlFor="name">{t('staffs.name')}</Label>
            <Input id="name" autoFocus {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="email">{t('staffs.email')}</Label>
              <Input id="email" type="email" {...form.register('email')} />
              {form.formState.errors.email && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone">{t('staffs.phone')}</Label>
              <Input id="phone" {...form.register('phone')} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="roleId">{t('staffs.role')}</Label>
              <select
                id="roleId"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('roleId')}
              >
                <option value="">{t('staffs.selectRole')}</option>
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
              <Label htmlFor="password">{t('staffs.password')}</Label>
              <Input id="password" type="password" {...form.register('password')} />
              {form.formState.errors.password && (
                <p className="text-xs text-[color:var(--color-destructive)]">
                  {form.formState.errors.password.message}
                </p>
              )}
            </div>
          </div>

          <StoreAccessSection
            allStores={form.watch('allStores')}
            storeIds={form.watch('storeIds') ?? []}
            onChange={(patch) => {
              if (patch.allStores !== undefined) {
                form.setValue('allStores', patch.allStores, { shouldDirty: true })
              }
              if (patch.storeIds !== undefined) {
                form.setValue('storeIds', patch.storeIds, { shouldDirty: true })
              }
            }}
          />

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

function EditStaffDialog({
  open,
  onOpenChange,
  existing,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  existing: StaffRow
}) {
  const { t } = useTranslation()
  const { data: roles } = useRoleOptions()
  const update = useUpdateStaff()

  const form = useForm<UpdateStaffFormData>({
    resolver: standardSchemaResolver(updateStaffSchema(t)),
    defaultValues: {
      name: '',
      phone: '',
      roleId: '',
      status: 'active',
      allStores: true,
      storeIds: [],
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        name: existing.name,
        phone: existing.phone,
        roleId: existing.roleId,
        status: (existing.status as UpdateStaffFormData['status']) ?? 'active',
        allStores: existing.allStores,
        storeIds: existing.storeIds ?? [],
      })
      update.reset()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, existing])

  const errorMessage = update.error ? ConnectError.from(update.error).rawMessage : null

  const onSubmit = form.handleSubmit(async (values) => {
    await update.mutateAsync({
      id: existing.id,
      member: {
        ...values,
        storeIds: values.allStores ? [] : values.storeIds,
      },
    })
    onOpenChange(false)
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('staffs.editStaff')}</DialogTitle>
        </DialogHeader>
        {errorMessage && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMessage}
          </div>
        )}
        <form onSubmit={onSubmit} className="space-y-4" noValidate>
          <div className="space-y-2">
            <Label>{t('staffs.email')}</Label>
            <Input value={existing.email} disabled readOnly />
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">{t('staffs.name')}</Label>
            <Input id="name" {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-[color:var(--color-destructive)]">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone">{t('staffs.phone')}</Label>
            <Input id="phone" {...form.register('phone')} />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="roleId">{t('staffs.role')}</Label>
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
              <Label htmlFor="status">{t('staffs.status')}</Label>
              <select
                id="status"
                className="h-10 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 text-sm"
                {...form.register('status')}
              >
                <option value="active">{t('staffs.statusActive')}</option>
                <option value="inactive">{t('staffs.statusInactive')}</option>
                <option value="suspended">{t('staffs.statusSuspended')}</option>
              </select>
            </div>
          </div>

          <StoreAccessSection
            allStores={form.watch('allStores')}
            storeIds={form.watch('storeIds') ?? []}
            onChange={(patch) => {
              if (patch.allStores !== undefined) {
                form.setValue('allStores', patch.allStores, { shouldDirty: true })
              }
              if (patch.storeIds !== undefined) {
                form.setValue('storeIds', patch.storeIds, { shouldDirty: true })
              }
            }}
          />

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

type StoreAccessPatch = { allStores?: boolean; storeIds?: string[] }

function StoreAccessSection({
  allStores,
  storeIds,
  onChange,
}: {
  allStores: boolean
  storeIds: string[]
  onChange: (patch: StoreAccessPatch) => void
}) {
  const { t } = useTranslation()
  const { data: stores, isLoading } = useStores()

  const toggleStore = (id: string) => {
    const next = storeIds.includes(id)
      ? storeIds.filter((s) => s !== id)
      : [...storeIds, id]
    onChange({ storeIds: next })
  }

  return (
    <div className="space-y-2 rounded-md border border-[color:var(--color-border)] p-3">
      <div className="text-sm font-medium">{t('staffs.storeAccess')}</div>

      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <input
          type="checkbox"
          className="h-4 w-4"
          checked={allStores}
          onChange={(e) => onChange({ allStores: e.target.checked })}
        />
        <span>{t('staffs.allStoresLabel')}</span>
      </label>

      {!allStores && (
        <div className="space-y-1 pl-6 pt-1">
          {isLoading && (
            <div className="text-xs text-[color:var(--color-muted-foreground)]">
              {t('common.loading')}
            </div>
          )}
          {!isLoading && (stores ?? []).length === 0 && (
            <div className="text-xs text-[color:var(--color-muted-foreground)]">
              {t('staffs.noStoresHint')}
            </div>
          )}
          {(stores ?? []).map((s) => (
            <label key={s.id} className="flex cursor-pointer items-center gap-2 text-sm">
              <input
                type="checkbox"
                className="h-4 w-4"
                checked={storeIds.includes(s.id)}
                onChange={() => toggleStore(s.id)}
              />
              <span>{s.name}</span>
            </label>
          ))}
          {!isLoading && (stores ?? []).length > 0 && storeIds.length === 0 && (
            <p className="pt-1 text-xs text-[color:var(--color-muted-foreground)]">
              {t('staffs.noStoresPickedHint')}
            </p>
          )}
        </div>
      )}
    </div>
  )
}
