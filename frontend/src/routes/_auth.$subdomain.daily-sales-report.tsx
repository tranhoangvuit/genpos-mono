import { Outlet, createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/$subdomain/daily-sales-report')({
  component: () => <Outlet />,
})
