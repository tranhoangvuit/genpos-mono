import { createFileRoute } from '@tanstack/react-router'

import { PurchaseOrderForm } from '@/features/inventory/PurchaseOrderForm'

export const Route = createFileRoute('/_auth/$subdomain/inventory/purchase-orders/new')({
  component: () => <PurchaseOrderForm />,
})
