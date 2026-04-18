import { createFileRoute, useParams } from '@tanstack/react-router'

import { StockTakeDetail } from '@/features/inventory/StockTakeDetail'

export const Route = createFileRoute('/_auth/inventory/stock-takes/$id')({
  component: StockTakeDetailRoute,
})

function StockTakeDetailRoute() {
  const { id } = useParams({ from: '/_auth/inventory/stock-takes/$id' })
  return <StockTakeDetail takeId={id} />
}
