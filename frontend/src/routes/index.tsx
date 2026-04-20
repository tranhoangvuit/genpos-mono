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
      const user = useAuthStore.getState().user
      if (user?.orgSlug) {
        void navigate({
          to: '/$subdomain/dashboard',
          params: { subdomain: user.orgSlug },
          replace: true,
        })
      } else {
        void navigate({ to: '/signin', replace: true })
      }
    })
  }, [navigate])
  return null
}
