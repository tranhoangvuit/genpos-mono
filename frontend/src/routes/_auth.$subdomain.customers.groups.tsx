import { createFileRoute } from '@tanstack/react-router'

import { CustomerGroupsPage } from '@/features/customers/CustomerGroupsPage'

export const Route = createFileRoute('/_auth/$subdomain/customers/groups')({
  component: CustomerGroupsPage,
})
