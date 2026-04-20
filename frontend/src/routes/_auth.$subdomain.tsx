import {
  Outlet,
  createFileRoute,
  useNavigate,
  useParams,
} from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { bootstrapAuth } from '@/shared/auth/bootstrap'
import { useAuthStore } from '@/shared/auth/store'
import { SyncProvider } from '@/shared/sync/provider'
import { SidebarProvider } from '@/shared/ui/sidebar'
import { TopNavbar } from '@/features/shell/TopNavbar'
import { AppSidebar } from '@/features/shell/AppSidebar'

export const Route = createFileRoute('/_auth/$subdomain')({
  component: AuthLayout,
})

function AuthLayout() {
  const user = useAuthStore((s) => s.user)
  const { subdomain } = useParams({ from: '/_auth/$subdomain' })
  const navigate = useNavigate()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    let cancelled = false
    void bootstrapAuth().finally(() => {
      if (cancelled) return
      const current = useAuthStore.getState().user
      if (!current) {
        void navigate({ to: '/signin', replace: true })
        return
      }
      if (current.orgSlug && subdomain !== current.orgSlug) {
        void navigate({
          to: '/$subdomain/dashboard',
          params: { subdomain: current.orgSlug },
          replace: true,
        })
        return
      }
      setChecked(true)
    })
    return () => {
      cancelled = true
    }
  }, [navigate, subdomain])

  if (!checked || !user) return null

  return (
    <SyncProvider>
      <SidebarProvider>
        <div className="flex min-h-svh w-full flex-col">
          <TopNavbar />
          <div className="flex flex-1 overflow-hidden">
            <AppSidebar />
            <main className="flex-1 overflow-y-auto p-6">
              <Outlet />
            </main>
          </div>
        </div>
      </SidebarProvider>
    </SyncProvider>
  )
}
