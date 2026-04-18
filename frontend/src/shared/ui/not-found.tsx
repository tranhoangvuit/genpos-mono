import { Link } from '@tanstack/react-router'

import { Button } from './button'

export function NotFound() {
  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
      <div>
        <h1 className="text-3xl font-semibold">404</h1>
        <p className="text-sm text-[color:var(--color-muted-foreground)]">
          Page not found.
        </p>
      </div>
      <Button asChild>
        <Link to="/">Go home</Link>
      </Button>
    </div>
  )
}
