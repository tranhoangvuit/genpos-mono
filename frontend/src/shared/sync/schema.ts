import { column, Schema, Table } from '@powersync/web'

const organizations = new Table({
  slug: column.text,
  name: column.text,
  currency: column.text,
  timezone: column.text,
  status: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const stores = new Table({
  org_id: column.text,
  name: column.text,
  address: column.text,
  phone: column.text,
  email: column.text,
  timezone: column.text,
  status: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const registers = new Table({
  org_id: column.text,
  store_id: column.text,
  name: column.text,
  status: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const roles = new Table({
  org_id: column.text,
  name: column.text,
  permissions: column.text,
  is_system: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const users = new Table({
  org_id: column.text,
  role_id: column.text,
  name: column.text,
  email: column.text,
  phone: column.text,
  status: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const categories = new Table({
  org_id: column.text,
  parent_id: column.text,
  name: column.text,
  sort_order: column.integer,
  color: column.text,
  image_url: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const products = new Table({
  org_id: column.text,
  category_id: column.text,
  name: column.text,
  description: column.text,
  image_url: column.text,
  is_active: column.integer,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const product_variants = new Table({
  org_id: column.text,
  product_id: column.text,
  name: column.text,
  sku: column.text,
  barcode: column.text,
  price: column.text,
  cost_price: column.text,
  track_stock: column.integer,
  is_active: column.integer,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const product_options = new Table({
  org_id: column.text,
  product_id: column.text,
  name: column.text,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const product_option_values = new Table({
  org_id: column.text,
  option_id: column.text,
  value: column.text,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const product_variant_option_values = new Table({
  org_id: column.text,
  variant_id: column.text,
  option_value_id: column.text,
  created_at: column.text,
})

const product_images = new Table({
  org_id: column.text,
  product_id: column.text,
  variant_id: column.text,
  url: column.text,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const tax_rates = new Table({
  org_id: column.text,
  name: column.text,
  rate: column.text,
  is_inclusive: column.integer,
  is_default: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const payment_methods = new Table({
  org_id: column.text,
  name: column.text,
  type: column.text,
  is_active: column.integer,
  sort_order: column.integer,
  created_at: column.text,
  updated_at: column.text,
})

const customers = new Table({
  org_id: column.text,
  name: column.text,
  email: column.text,
  phone: column.text,
  notes: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const customer_groups = new Table({
  org_id: column.text,
  name: column.text,
  description: column.text,
  discount_type: column.text,
  discount_value: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const customer_group_members = new Table({
  org_id: column.text,
  group_id: column.text,
  customer_id: column.text,
  created_at: column.text,
})

const suppliers = new Table({
  org_id: column.text,
  name: column.text,
  contact_name: column.text,
  email: column.text,
  phone: column.text,
  address: column.text,
  notes: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const purchase_orders = new Table({
  org_id: column.text,
  store_id: column.text,
  user_id: column.text,
  po_number: column.text,
  supplier_name: column.text,
  status: column.text,
  notes: column.text,
  expected_at: column.text,
  received_at: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const purchase_order_items = new Table({
  org_id: column.text,
  purchase_order_id: column.text,
  variant_id: column.text,
  quantity_ordered: column.text,
  quantity_received: column.text,
  cost_price: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const stock_movements = new Table({
  org_id: column.text,
  store_id: column.text,
  variant_id: column.text,
  direction: column.text,
  quantity: column.text,
  movement_type: column.text,
  reference_type: column.text,
  reference_id: column.text,
  created_at: column.text,
})

const stock_takes = new Table({
  org_id: column.text,
  store_id: column.text,
  user_id: column.text,
  status: column.text,
  notes: column.text,
  completed_at: column.text,
  created_at: column.text,
  updated_at: column.text,
})

const stock_take_items = new Table({
  org_id: column.text,
  stock_take_id: column.text,
  variant_id: column.text,
  expected_qty: column.text,
  counted_qty: column.text,
  created_at: column.text,
  updated_at: column.text,
})

export const appSchema = new Schema({
  organizations,
  stores,
  registers,
  roles,
  users,
  categories,
  products,
  product_variants,
  product_options,
  product_option_values,
  product_variant_option_values,
  product_images,
  tax_rates,
  payment_methods,
  customers,
  customer_groups,
  customer_group_members,
  suppliers,
  purchase_orders,
  purchase_order_items,
  stock_movements,
  stock_takes,
  stock_take_items,
})
