import { createFileRoute } from '@tanstack/react-router'

import { AuthLayout } from '@/features/auth/AuthLayout'
import { LoginCard } from '@/features/auth/LoginCard'

export const Route = createFileRoute('/_guest/signin')({
  component: SignInPage,
})

function SignInPage() {
  return (
    <AuthLayout>
      <LoginCard />
    </AuthLayout>
  )
}
