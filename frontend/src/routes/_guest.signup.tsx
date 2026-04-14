import { createFileRoute } from '@tanstack/react-router'

import { AuthLayout } from '@/features/auth/AuthLayout'
import { RegisterCard } from '@/features/auth/RegisterCard'

export const Route = createFileRoute('/_guest/signup')({
  component: SignUpPage,
})

function SignUpPage() {
  return (
    <AuthLayout>
      <RegisterCard />
    </AuthLayout>
  )
}
