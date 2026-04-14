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
      if (useAuthStore.getState().user) {
        void navigate({ to: '/dashboard', replace: true })
      }
    })
    return () => {
      cancelled = true
    }
  }, [navigate])

  return <Outlet />
}
