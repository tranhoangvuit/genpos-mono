import type { TFunction } from 'i18next'
import { z } from 'zod'

export function customerSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('customers.validation.nameRequired')).max(120),
    email: z
      .string()
      .max(200)
      .refine(
        (v) => v === '' || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
        t('customers.validation.emailInvalid'),
      ),
    phone: z.string().max(40),
    notes: z.string().max(2000),
    groupIds: z.array(z.string()),
  })
}
export type CustomerFormData = z.infer<ReturnType<typeof customerSchema>>

export function customerGroupSchema(t: TFunction) {
  return z
    .object({
      name: z.string().min(1, t('customers.validation.groupNameRequired')).max(120),
      description: z.string().max(500),
      discountType: z.enum(['', 'percentage', 'fixed']),
      discountValue: z.string(),
    })
    .refine(
      (v) =>
        v.discountValue === '' ||
        (!Number.isNaN(Number(v.discountValue)) && Number(v.discountValue) >= 0),
      {
        message: t('customers.validation.discountInvalid'),
        path: ['discountValue'],
      },
    )
    .refine(
      (v) => !(v.discountValue !== '' && v.discountType === ''),
      {
        message: t('customers.validation.discountTypeRequired'),
        path: ['discountType'],
      },
    )
}
export type CustomerGroupFormData = z.infer<ReturnType<typeof customerGroupSchema>>
