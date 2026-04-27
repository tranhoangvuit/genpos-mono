import { createFileRoute } from '@tanstack/react-router'

import { StaffsPage } from '@/features/settings/staffs/StaffsPage'

export const Route = createFileRoute('/_auth/$subdomain/settings/staffs')({
  component: StaffsPage,
})
