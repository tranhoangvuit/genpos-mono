import type { TFunction } from 'i18next'
import { z } from 'zod'

export function categorySchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('catalog.validation.nameRequired')).max(80),
    parentId: z.string(),
    color: z.string(),
    sortOrder: z.number().int(),
  })
}
export type CategoryFormData = z.infer<ReturnType<typeof categorySchema>>

export function variantSchema(t: TFunction) {
  return z.object({
    name: z.string(),
    sku: z.string(),
    barcode: z.string(),
    price: z
      .string()
      .min(1, t('catalog.validation.priceRequired'))
      .refine((v) => !Number.isNaN(Number(v)), t('catalog.validation.priceInvalid')),
    costPrice: z.string(),
    trackStock: z.boolean(),
    isActive: z.boolean(),
    sortOrder: z.number().int(),
    optionValues: z.array(z.string()),
  })
}

export function optionSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('catalog.validation.optionNameRequired')),
    values: z
      .array(z.string().min(1))
      .min(1, t('catalog.validation.optionValuesRequired')),
  })
}

export function productSchema(t: TFunction) {
  return z
    .object({
      name: z.string().min(1, t('catalog.validation.nameRequired')).max(120),
      description: z.string(),
      categoryId: z.string(),
      isActive: z.boolean(),
      sortOrder: z.number().int(),
      hasVariants: z.boolean(),
      options: z.array(optionSchema(t)),
      variants: z.array(variantSchema(t)).min(1, t('catalog.validation.variantRequired')),
      images: z.array(z.object({ url: z.string(), sortOrder: z.number().int() })),
    })
    .refine((v) => !v.hasVariants || v.options.length > 0, {
      message: t('catalog.validation.optionsRequired'),
      path: ['options'],
    })
}
export type ProductFormData = z.infer<ReturnType<typeof productSchema>>
