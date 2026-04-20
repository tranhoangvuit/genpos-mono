import { createFileRoute } from '@tanstack/react-router'

import { SuppliersPage } from '@/features/inventory/SuppliersPage'

export const Route = createFileRoute('/_auth/$subdomain/inventory/suppliers')({
  component: SuppliersPage,
})
