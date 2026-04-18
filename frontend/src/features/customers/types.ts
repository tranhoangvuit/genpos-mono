import type { Customer, CustomerGroup, CustomerListItem } from '@/gen/genpos/v1/customer_pb'

export type CustomerRow = Customer
export type CustomerGroupRow = CustomerGroup
export type CustomerListRow = CustomerListItem

export type CustomerFormValues = {
  name: string
  email: string
  phone: string
  notes: string
  groupIds: string[]
}

export type CustomerGroupFormValues = {
  name: string
  description: string
  discountType: '' | 'percentage' | 'fixed'
  discountValue: string
}
