import type { TFunction } from 'i18next'
import { z } from 'zod'

const storeAccessShape = {
  allStores: z.boolean(),
  storeIds: z.array(z.string()),
} as const

export function createStaffSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('staffs.validation.nameRequired')).max(120),
    email: z
      .string()
      .min(1, t('staffs.validation.emailRequired'))
      .max(200)
      .refine(
        (v) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
        t('staffs.validation.emailInvalid'),
      ),
    phone: z.string().max(40),
    roleId: z.string().min(1, t('staffs.validation.roleRequired')),
    password: z.string().min(8, t('staffs.validation.passwordMin')).max(200),
    ...storeAccessShape,
  })
}

export function updateStaffSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('staffs.validation.nameRequired')).max(120),
    phone: z.string().max(40),
    roleId: z.string().min(1, t('staffs.validation.roleRequired')),
    status: z.enum(['active', 'inactive', 'suspended']),
    ...storeAccessShape,
  })
}

export type CreateStaffFormData = z.infer<ReturnType<typeof createStaffSchema>>
export type UpdateStaffFormData = z.infer<ReturnType<typeof updateStaffSchema>>
