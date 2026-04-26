import { createFileRoute } from '@tanstack/react-router'

import { StockInForm } from '@/features/inventory/StockInForm'

export const Route = createFileRoute('/_auth/$subdomain/inventory/stock-ins/new')({
  component: () => <StockInForm />,
})
