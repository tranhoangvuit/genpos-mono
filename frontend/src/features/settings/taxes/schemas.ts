import type { TFunction } from 'i18next'
import { z } from 'zod'

export function taxRateSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('taxes.validation.nameRequired')).max(120),
    // User inputs percent (e.g. "10" → 10%). Converted to fraction on save.
    ratePercent: z
      .string()
      .refine(
        (v) => v !== '' && !Number.isNaN(Number(v)) && Number(v) >= 0 && Number(v) < 100,
        t('taxes.validation.rateInvalid'),
      ),
    isInclusive: z.boolean(),
    isDefault: z.boolean(),
  })
}

export type TaxRateFormData = z.infer<ReturnType<typeof taxRateSchema>>

export function percentToFraction(percent: string): string {
  const n = Number(percent)
  if (!Number.isFinite(n)) return '0'
  return (n / 100).toFixed(4)
}

export function fractionToPercent(fraction: string): string {
  const n = Number(fraction)
  if (!Number.isFinite(n)) return '0'
  // Strip trailing zeros on display (0.1000 → 10).
  return String(Math.round(n * 10000) / 100)
}
