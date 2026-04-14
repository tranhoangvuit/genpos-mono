import type { ReactNode } from 'react'

import { BrandPanel } from './BrandPanel'

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <BrandPanel />
      <div className="flex items-center justify-center p-6 lg:p-10">
        <div className="w-full max-w-md">{children}</div>
      </div>
    </div>
  )
}
