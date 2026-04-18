import type { TFunction } from 'i18next'
import { z } from 'zod'

export function supplierSchema(t: TFunction) {
  return z.object({
    name: z.string().min(1, t('inventory.validation.nameRequired')).max(120),
    contactName: z.string().max(120),
    email: z
      .string()
      .max(200)
      .refine(
        (v) => v === '' || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v),
        t('inventory.validation.emailInvalid'),
      ),
    phone: z.string().max(40),
    address: z.string().max(500),
    notes: z.string().max(2000),
  })
}
export type SupplierFormData = z.infer<ReturnType<typeof supplierSchema>>

export function purchaseOrderItemSchema(t: TFunction) {
  return z.object({
    variantId: z.string().min(1, t('inventory.validation.variantRequired')),
    variantLabel: z.string(),
    quantityOrdered: z
      .string()
      .refine(
        (v) => Number(v) > 0,
        t('inventory.validation.quantityPositive'),
      ),
    costPrice: z
      .string()
      .refine(
        (v) => v === '' || (!Number.isNaN(Number(v)) && Number(v) >= 0),
        t('inventory.validation.costInvalid'),
      ),
  })
}

export function purchaseOrderSchema(t: TFunction) {
  return z.object({
    storeId: z.string().min(1, t('inventory.validation.storeRequired')),
    supplierName: z.string().max(200),
    notes: z.string().max(2000),
    expectedAt: z.string(),
    items: z
      .array(purchaseOrderItemSchema(t))
      .min(1, t('inventory.validation.itemsRequired')),
  })
}
export type PurchaseOrderFormData = z.infer<ReturnType<typeof purchaseOrderSchema>>
