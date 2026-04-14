import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { Link } from '@tanstack/react-router'
import { ConnectError } from '@connectrpc/connect'
import { Eye, EyeOff } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { useSignUp } from '@/shared/auth/hooks'
import { signUpSchema, type SignUpValues } from '@/shared/auth/schemas'

export function RegisterCard() {
  const { t } = useTranslation()
  const signUp = useSignUp()
  const [showPassword, setShowPassword] = useState(false)

  const form = useForm<SignUpValues>({
    resolver: standardSchemaResolver(signUpSchema(t)),
    defaultValues: { domain: '', email: '', password: '' },
  })

  const onSubmit = form.handleSubmit((values) => {
    signUp.mutate(values)
  })

  const serverMessage = signUp.error
    ? ConnectError.from(signUp.error).rawMessage || t('common.unexpectedError')
    : null

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-2xl font-semibold tracking-tight">
          {t('auth.signUpTitle')}
        </h2>
        <p className="text-sm text-[color:var(--color-muted-foreground)]">
          {t('auth.signUpSubtitle')}
        </p>
      </div>

      {serverMessage && (
        <div
          role="alert"
          className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]"
        >
          {serverMessage}
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-4" noValidate>
        <div className="space-y-2">
          <Label htmlFor="domain">{t('auth.businessDomain')}</Label>
          <Input
            id="domain"
            type="text"
            autoComplete="organization"
            placeholder="mybusiness"
            aria-invalid={form.formState.errors.domain ? true : undefined}
            {...form.register('domain')}
          />
          <p className="text-xs text-[color:var(--color-muted-foreground)]">
            {t('auth.businessDomainHint')}
          </p>
          {form.formState.errors.domain && (
            <p className="text-xs text-[color:var(--color-destructive)]">
              {form.formState.errors.domain.message}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="email">{t('auth.email')}</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            aria-invalid={form.formState.errors.email ? true : undefined}
            {...form.register('email')}
          />
          {form.formState.errors.email && (
            <p className="text-xs text-[color:var(--color-destructive)]">
              {form.formState.errors.email.message}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">{t('auth.password')}</Label>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? 'text' : 'password'}
              autoComplete="new-password"
              aria-invalid={form.formState.errors.password ? true : undefined}
              className="pr-10"
              {...form.register('password')}
            />
            <button
              type="button"
              onClick={() => setShowPassword((s) => !s)}
              className="absolute inset-y-0 right-0 flex items-center px-3 text-[color:var(--color-muted-foreground)] hover:text-[color:var(--color-foreground)]"
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              tabIndex={-1}
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </button>
          </div>
          {form.formState.errors.password && (
            <p className="text-xs text-[color:var(--color-destructive)]">
              {form.formState.errors.password.message}
            </p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={signUp.isPending}>
          {signUp.isPending ? t('auth.creatingAccount') : t('auth.signUp')}
        </Button>
      </form>

      <p className="text-center text-sm text-[color:var(--color-muted-foreground)]">
        {t('auth.haveAccount')}{' '}
        <Link
          to="/signin"
          className="font-medium text-[color:var(--color-foreground)] underline-offset-4 hover:underline"
        >
          {t('auth.signIn')}
        </Link>
      </p>
    </div>
  )
}
