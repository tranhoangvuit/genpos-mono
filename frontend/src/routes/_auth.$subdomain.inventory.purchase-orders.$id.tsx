import { createFileRoute, useParams } from '@tanstack/react-router'

import { PurchaseOrderDetail } from '@/features/inventory/PurchaseOrderDetail'

export const Route = createFileRoute('/_auth/$subdomain/inventory/purchase-orders/$id')({
  component: PurchaseOrderDetailRoute,
})

function PurchaseOrderDetailRoute() {
  const { id } = useParams({ from: '/_auth/$subdomain/inventory/purchase-orders/$id' })
  return <PurchaseOrderDetail poId={id} />
}
