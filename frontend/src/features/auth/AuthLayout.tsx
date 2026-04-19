import type { ReactNode } from 'react'
import { Link } from '@tanstack/react-router'

import { BrandPanel } from './BrandPanel'

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="grid min-h-svh lg:grid-cols-[1.1fr_1fr]">
      <BrandPanel />
      <div className="relative flex items-start justify-center overflow-y-auto bg-[hsl(210_40%_98%)] p-6 pt-20 lg:items-center lg:p-10 lg:pt-10">
        <div className="absolute right-7 top-6 text-[13px] text-[hsl(215.4_16.3%_46.9%)]">
          New to GenPos?{' '}
          <Link
            to="/signup"
            className="ml-1 font-semibold text-[hsl(221_83%_53%)] hover:underline"
          >
            Start a 14-day trial
          </Link>
        </div>
        <div className="w-full max-w-[400px]">{children}</div>
      </div>
    </div>
  )
}
