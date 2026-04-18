import { createFileRoute } from '@tanstack/react-router'

import { StockTakesPage } from '@/features/inventory/StockTakesPage'

export const Route = createFileRoute('/_auth/inventory/stock-takes/')({
  component: StockTakesPage,
})
