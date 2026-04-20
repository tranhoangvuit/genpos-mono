import { createFileRoute } from '@tanstack/react-router'

import { TaxRatesPage } from '@/features/settings/taxes/TaxRatesPage'

export const Route = createFileRoute('/_auth/$subdomain/settings/taxes')({
  component: TaxRatesPage,
})
