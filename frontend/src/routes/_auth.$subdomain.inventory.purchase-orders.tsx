import { Outlet, createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/$subdomain/inventory/purchase-orders')({
  component: () => <Outlet />,
})
