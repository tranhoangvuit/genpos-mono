import { useCallback, useEffect, useRef } from 'react'
import { Bell, Download, LogOut, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Link } from '@tanstack/react-router'

import { useSignOut } from '@/shared/auth/hooks'
import { useAuthStore } from '@/shared/auth/store'
import { Button } from '@/shared/ui/button'
import { SidebarTrigger } from '@/shared/ui/sidebar'
import { Separator } from '@/shared/ui/separator'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'

export function TopNavbar() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const subdomain = user?.orgSlug ?? ''
  const signOut = useSignOut()
  const searchRef = useRef<HTMLInputElement>(null)

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      searchRef.current?.focus()
    }
  }, [])

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const displayName = user?.name || user?.email || ''
  const initials = displayName
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <header className="flex h-12 shrink-0 items-center gap-2 border-b bg-[color:var(--color-card)] px-4">
      <SidebarTrigger className="-ml-1" />
      <Separator orientation="vertical" className="mr-2 h-4" />

      <span className="text-sm font-semibold tracking-tight">
        {t('auth.brand')}
      </span>

      <div className="mx-auto flex w-full max-w-sm items-center">
        <div className="relative w-full">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-[color:var(--color-muted-foreground)]" />
          <input
            ref={searchRef}
            type="text"
            placeholder={t('nav.searchPlaceholder')}
            className="h-8 w-full rounded-md border border-[color:var(--color-input)] bg-[color:var(--color-background)] pl-8 pr-12 text-sm placeholder:text-[color:var(--color-muted-foreground)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[color:var(--color-ring)]"
          />
          <kbd className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 rounded border border-[color:var(--color-border)] bg-[color:var(--color-muted)] px-1.5 text-[10px] font-medium text-[color:var(--color-muted-foreground)]">
            ⌘K
          </kbd>
        </div>
      </div>

      <div className="flex items-center gap-1">
        <Button asChild variant="outline" size="sm" className="h-8 gap-1.5 px-2.5">
          <Link to="/$subdomain/downloads" params={{ subdomain }}>
            <Download className="h-4 w-4" />
            <span className="hidden sm:inline-block">
              {t('nav.downloads')}
            </span>
          </Link>
        </Button>

        <Button variant="ghost" size="icon" className="h-8 w-8">
          <Bell className="h-4 w-4" />
          <span className="sr-only">{t('nav.notifications')}</span>
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className="h-8 gap-2 px-2 text-sm font-normal"
            >
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-[color:var(--color-primary)] text-[10px] font-medium text-[color:var(--color-primary-foreground)]">
                {initials}
              </span>
              <span className="hidden sm:inline-block">{displayName}</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col gap-1">
                <p className="text-sm font-medium leading-none">{displayName}</p>
                <p className="text-xs leading-none text-[color:var(--color-muted-foreground)]">
                  {user?.email}
                </p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => signOut.mutate()}
              disabled={signOut.isPending}
            >
              <LogOut className="mr-2 h-4 w-4" />
              {t('dashboard.signOut')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
