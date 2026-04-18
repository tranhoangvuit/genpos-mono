import type { PaymentMethod } from '@/gen/genpos/v1/payment_pb'

export type PaymentMethodRow = PaymentMethod

export const PAYMENT_METHOD_TYPES = [
  'cash',
  'card',
  'mobile',
  'bank_transfer',
  'voucher',
  'other',
] as const
