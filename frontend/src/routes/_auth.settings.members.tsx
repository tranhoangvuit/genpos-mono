import { createFileRoute } from '@tanstack/react-router'

import { MembersPage } from '@/features/settings/members/MembersPage'

export const Route = createFileRoute('/_auth/settings/members')({
  component: MembersPage,
})
