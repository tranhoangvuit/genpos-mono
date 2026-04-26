import { createFileRoute } from '@tanstack/react-router'

import { StocksPage } from '@/features/inventory/StocksPage'

export const Route = createFileRoute('/_auth/$subdomain/inventory/stocks/')({
  component: StocksPage,
})
