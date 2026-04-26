import { createFileRoute } from '@tanstack/react-router'

import { DownloadsPage } from '@/features/downloads/DownloadsPage'

export const Route = createFileRoute('/_auth/$subdomain/downloads')({
  component: DownloadsPage,
  head: () => ({
    meta: [{ title: 'Get the GenPos app' }],
  }),
})
