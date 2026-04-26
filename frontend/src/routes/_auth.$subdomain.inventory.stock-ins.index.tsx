import { createFileRoute } from '@tanstack/react-router'

import { StockInsPage } from '@/features/inventory/StockInsPage'

export const Route = createFileRoute('/_auth/$subdomain/inventory/stock-ins/')({
  component: StockInsPage,
})
