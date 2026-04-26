import { Outlet, createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/$subdomain/inventory/stock-ins')({
  component: () => <Outlet />,
})
