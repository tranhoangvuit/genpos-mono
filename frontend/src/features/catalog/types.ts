import type { Category, ProductListItem } from '@/gen/genpos/v1/catalog_pb'

export type CategoryRow = Category
export type ProductListRow = ProductListItem

// Form types (mirror proto ProductInput)
export type OptionFormValue = {
  name: string
  values: string[]
}

export type VariantFormValue = {
  name: string
  sku: string
  barcode: string
  price: string
  costPrice: string
  trackStock: boolean
  isActive: boolean
  sortOrder: number
  optionValues: string[]
}

export type ProductImageFormValue = {
  url: string
  sortOrder: number
}

export type ProductFormValues = {
  name: string
  description: string
  categoryId: string
  isActive: boolean
  sortOrder: number
  hasVariants: boolean
  options: OptionFormValue[]
  variants: VariantFormValue[]
  images: ProductImageFormValue[]
}

export type CategoryFormValues = {
  name: string
  parentId: string
  color: string
  sortOrder: number
}
