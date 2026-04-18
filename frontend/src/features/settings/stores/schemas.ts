import type { TFunction } from 'i18next'
import { z } from 'zod'

export function storeSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('stores.validation.nameRequired')).max(120),
    address: z.string().max(500),
    phone: z.string().max(40),
    email: z
      .string()
      .max(200)
      .refine(
        (v) => v === '' || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
        t('stores.validation.emailInvalid'),
      ),
    timezone: z.string().max(60),
    status: z.enum(['', 'active', 'inactive', 'closed']),
  })
}

export type StoreFormData = z.infer<ReturnType<typeof storeSchema>>
