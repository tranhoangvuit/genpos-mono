import {
  Outlet,
  createFileRoute,
  useNavigate,
} from '@tanstack/react-router'
import { useEffect } from 'react'

import { bootstrapAuth } from '@/shared/auth/bootstrap'
import { useAuthStore } from '@/shared/auth/store'

export const Route = createFileRoute('/_guest')({
  component: GuestLayout,
})

function GuestLayout() {
  const navigate = useNavigate()

  useEffect(() => {
    let cancelled = false
    void bootstrapAuth().finally(() => {
      if (cancelled) return
      const user = useAuthStore.getState().user
      if (user?.orgSlug) {
        void navigate({
          to: '/$subdomain/dashboard',
          params: { subdomain: user.orgSlug },
          replace: true,
        })
      }
    })
    return () => {
      cancelled = true
    }
  }, [navigate])

  return <Outlet />
}
