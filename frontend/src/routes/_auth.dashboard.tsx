import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'

export const Route = createFileRoute('/_auth/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)

  if (!user) return null

  const displayName = user.name || user.email

  return (
    <div className="space-y-2">
      <h1 className="text-2xl font-semibold tracking-tight">
        {t('dashboard.welcome', { name: displayName })}
      </h1>
      <p className="text-sm text-[color:var(--color-muted-foreground)]">
        {t('dashboard.signedInAs', {
          email: user.email,
          domain: user.orgSlug,
        })}
      </p>
    </div>
  )
}
