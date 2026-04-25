import { Link, createFileRoute } from '@tanstack/react-router'

import { AuthLayout } from '@/features/auth/AuthLayout'
import { RegisterCard } from '@/features/auth/RegisterCard'

export const Route = createFileRoute('/_guest/signup')({
  component: SignUpPage,
})

function SignUpPage() {
  return (
    <AuthLayout
      topRight={
        <>
          Already have an account?
          <Link
            to="/signin"
            className="ml-1 font-semibold text-[hsl(221_83%_53%)] hover:underline"
          >
            Sign in
          </Link>
        </>
      }
    >
      <RegisterCard />
    </AuthLayout>
  )
}
