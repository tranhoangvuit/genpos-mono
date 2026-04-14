import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

import { bootstrapAuth } from '@/shared/auth/bootstrap'
import { useAuthStore } from '@/shared/auth/store'

export const Route = createFileRoute('/')({
  component: IndexRedirect,
})

function IndexRedirect() {
  const navigate = useNavigate()
  useEffect(() => {
    void bootstrapAuth().finally(() => {
      void navigate({
        to: useAuthStore.getState().user ? '/dashboard' : '/signin',
        replace: true,
      })
    })
  }, [navigate])
  return null
}
