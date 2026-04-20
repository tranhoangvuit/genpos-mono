import { createFileRoute } from '@tanstack/react-router'

import { StockTakesPage } from '@/features/inventory/StockTakesPage'

export const Route = createFileRoute('/_auth/$subdomain/inventory/stock-takes/')({
  component: StockTakesPage,
})
