import { createFileRoute, useParams } from '@tanstack/react-router'

import { OrderDetailPage } from '@/features/reports/OrderDetailPage'

export const Route = createFileRoute('/_auth/daily-sales-report/$orderId')({
  component: OrderDetailRoute,
})

function OrderDetailRoute() {
  const { orderId } = useParams({ from: '/_auth/daily-sales-report/$orderId' })
  return <OrderDetailPage orderId={orderId} />
}
