import type { TFunction } from 'i18next'
import { z } from 'zod'

export function createMemberSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('members.validation.nameRequired')).max(120),
    email: z
      .string()
      .min(1, t('members.validation.emailRequired'))
      .max(200)
      .refine(
        (v) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
        t('members.validation.emailInvalid'),
      ),
    phone: z.string().max(40),
    roleId: z.string().min(1, t('members.validation.roleRequired')),
    password: z.string().min(8, t('members.validation.passwordMin')).max(200),
  })
}

export function updateMemberSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('members.validation.nameRequired')).max(120),
    phone: z.string().max(40),
    roleId: z.string().min(1, t('members.validation.roleRequired')),
    status: z.enum(['active', 'inactive', 'suspended']),
  })
}

export type CreateMemberFormData = z.infer<ReturnType<typeof createMemberSchema>>
export type UpdateMemberFormData = z.infer<ReturnType<typeof updateMemberSchema>>
