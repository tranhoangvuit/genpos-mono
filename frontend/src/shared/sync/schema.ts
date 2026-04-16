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

export const appSchema = new Schema({
  organizations,
  stores,
  registers,
  roles,
  users,
  categories,
  products,
})
