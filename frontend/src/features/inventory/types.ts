import type {
  PurchaseOrder,
  PurchaseOrderItem,
  PurchaseOrderListItem,
  StoreRef,
  Supplier,
  VariantPickerItem,
} from '@/gen/genpos/v1/inventory_pb'
import type { StockTake, StockTakeItem, StockTakeListItem } from '@/gen/genpos/v1/stock_take_pb'

export type SupplierRow = Supplier
export type PurchaseOrderRow = PurchaseOrder
export type PurchaseOrderItemRow = PurchaseOrderItem
export type PurchaseOrderListRow = PurchaseOrderListItem
export type VariantPickerRow = VariantPickerItem
export type StoreRow = StoreRef

export type StockTakeRow = StockTake
export type StockTakeItemRow = StockTakeItem
export type StockTakeListRow = StockTakeListItem

export type SupplierFormValues = {
  name: string
  contactName: string
  email: string
  phone: string
  address: string
  notes: string
}

export type PurchaseOrderFormItem = {
  variantId: string
  variantLabel: string
  quantityOrdered: string
  costPrice: string
}

export type PurchaseOrderFormValues = {
  storeId: string
  supplierName: string
  notes: string
  expectedAt: string
  items: PurchaseOrderFormItem[]
}
