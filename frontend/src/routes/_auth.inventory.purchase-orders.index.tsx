import { createFileRoute } from '@tanstack/react-router'

import { PurchaseOrdersPage } from '@/features/inventory/PurchaseOrdersPage'

export const Route = createFileRoute('/_auth/inventory/purchase-orders/')({
  component: PurchaseOrdersPage,
})
