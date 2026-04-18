import type { TFunction } from 'i18next'
import { z } from 'zod'

export function paymentMethodSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('payments.validation.nameRequired')).max(120),
    type: z.enum(['cash', 'card', 'mobile', 'bank_transfer', 'voucher', 'other']),
    isActive: z.boolean(),
    sortOrder: z.number().int().min(0).max(9999),
  })
}

export type PaymentMethodFormData = z.infer<ReturnType<typeof paymentMethodSchema>>
