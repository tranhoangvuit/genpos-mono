import { useTranslation } from 'react-i18next'

function StorefrontMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}
    >
      <path d="M4 12 6.5 5h19L28 12" />
      <path d="M5 12v15h22V12" />
      <path d="M13 27v-8h6v8" />
      <path d="M4 12h24" />
    </svg>
  )
}

export function BrandPanel() {
  const { t } = useTranslation()
  const year = new Date().getFullYear()

  return (
    <div className="relative hidden flex-col justify-between overflow-hidden bg-[color:var(--color-primary)] p-10 text-[color:var(--color-primary-foreground)] lg:flex">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 bg-gradient-to-br from-white/8 via-transparent to-transparent"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-24 -top-24 h-72 w-72 rounded-full bg-white/5 blur-3xl"
      />

      <div className="relative flex items-center gap-3">
        <StorefrontMark className="h-9 w-9" />
        <span className="text-xl font-semibold tracking-tight">
          {t('auth.brand')}
        </span>
      </div>

      <div className="relative max-w-md">
        <h1 className="text-4xl font-semibold leading-tight tracking-tight">
          {t('auth.tagline')}
        </h1>
        <p className="mt-4 text-base text-white/70">{t('auth.subTagline')}</p>
      </div>

      <p className="relative text-sm text-white/50">
        {t('auth.copyright', { year })}
      </p>
    </div>
  )
}
