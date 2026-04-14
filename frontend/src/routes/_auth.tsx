import {
  Outlet,
  createFileRoute,
  useNavigate,
} from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { bootstrapAuth } from '@/shared/auth/bootstrap'
import { useAuthStore } from '@/shared/auth/store'

export const Route = createFileRoute('/_auth')({
  component: AuthLayout,
})

function AuthLayout() {
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    let cancelled = false
    void bootstrapAuth().finally(() => {
      if (cancelled) return
      if (!useAuthStore.getState().user) {
        void navigate({ to: '/signin', replace: true })
      } else {
        setChecked(true)
      }
    })
    return () => {
      cancelled = true
    }
  }, [navigate])

  if (!checked || !user) return null
  return <Outlet />
}
