import { createFileRoute } from '@tanstack/react-router'

import { CustomerGroupsPage } from '@/features/customers/CustomerGroupsPage'

export const Route = createFileRoute('/_auth/customers/groups')({
  component: CustomerGroupsPage,
})
