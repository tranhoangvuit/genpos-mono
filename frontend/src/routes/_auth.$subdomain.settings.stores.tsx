import { createFileRoute } from '@tanstack/react-router'

import { StoresPage } from '@/features/settings/stores/StoresPage'

export const Route = createFileRoute('/_auth/$subdomain/settings/stores')({
  component: StoresPage,
})
