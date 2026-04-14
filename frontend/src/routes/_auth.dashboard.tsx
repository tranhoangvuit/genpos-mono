import { createFileRoute } from '@tanstack/react-router'
import { LogOut } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useSignOut } from '@/shared/auth/hooks'
import { useAuthStore } from '@/shared/auth/store'
import { Button } from '@/shared/ui/button'

export const Route = createFileRoute('/_auth/dashboard')({
  component: DashboardPage,
})

function DashboardPage() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const signOut = useSignOut()

  if (!user) return null

  const displayName = user.name || user.email

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex items-center justify-between border-b border-[color:var(--color-border)] bg-[color:var(--color-card)] px-6 py-4">
        <span className="text-sm font-semibold tracking-tight">
          {t('auth.brand')}
        </span>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => signOut.mutate()}
          disabled={signOut.isPending}
        >
          <LogOut className="mr-2 h-4 w-4" />
          {t('dashboard.signOut')}
        </Button>
      </header>

      <main className="flex flex-1 items-center justify-center p-6">
        <div className="w-full max-w-md space-y-2 text-center">
          <h1 className="text-3xl font-semibold tracking-tight">
            {t('dashboard.welcome', { name: displayName })}
          </h1>
          <p className="text-sm text-[color:var(--color-muted-foreground)]">
            {t('dashboard.signedInAs', {
              email: user.email,
              domain: user.orgSlug,
            })}
          </p>
        </div>
      </main>
    </div>
  )
}
