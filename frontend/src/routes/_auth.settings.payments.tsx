import { createFileRoute } from '@tanstack/react-router'

import { PaymentMethodsPage } from '@/features/settings/payments/PaymentMethodsPage'

export const Route = createFileRoute('/_auth/settings/payments')({
  component: PaymentMethodsPage,
})
