import { createFileRoute } from '@tanstack/react-router'

import { SuppliersPage } from '@/features/inventory/SuppliersPage'

export const Route = createFileRoute('/_auth/inventory/suppliers')({
  component: SuppliersPage,
})
