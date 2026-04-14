import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { Link } from '@tanstack/react-router'
import { Code, ConnectError } from '@connectrpc/connect'
import { Eye, EyeOff } from 'lucide-react'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { useSignIn } from '@/shared/auth/hooks'
import { signInSchema, type SignInValues } from '@/shared/auth/schemas'

export function LoginCard() {
  const { t } = useTranslation()
  const signIn = useSignIn()
  const [showPassword, setShowPassword] = useState(false)

  const form = useForm<SignInValues>({
    resolver: standardSchemaResolver(signInSchema(t)),
    defaultValues: { email: '', password: '', rememberMe: false },
  })

  const onSubmit = form.handleSubmit((values) => {
    signIn.mutate(values)
  })

  const serverMessage = signIn.error
    ? humanizeSignInError(signIn.error, t)
    : null

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-2xl font-semibold tracking-tight">
          {t('auth.signIn')}
        </h2>
        <p className="text-sm text-[color:var(--color-muted-foreground)]">
          {t('auth.signInSubtitle')}
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
              autoComplete="current-password"
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

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Controller
              control={form.control}
              name="rememberMe"
              render={({ field }) => (
                <Checkbox
                  id="rememberMe"
                  checked={field.value}
                  onCheckedChange={(checked) =>
                    field.onChange(checked === true)
                  }
                />
              )}
            />
            <Label htmlFor="rememberMe" className="cursor-pointer">
              {t('auth.rememberMe')}
            </Label>
          </div>
          <a
            href="#"
            onClick={(e) => e.preventDefault()}
            className="text-sm text-[color:var(--color-muted-foreground)] hover:text-[color:var(--color-foreground)]"
          >
            {t('auth.forgotPassword')}
          </a>
        </div>

        <Button type="submit" className="w-full" disabled={signIn.isPending}>
          {signIn.isPending ? t('auth.signingIn') : t('auth.signIn')}
        </Button>
      </form>

      <p className="text-center text-sm text-[color:var(--color-muted-foreground)]">
        {t('auth.noAccount')}{' '}
        <Link
          to="/signup"
          className="font-medium text-[color:var(--color-foreground)] underline-offset-4 hover:underline"
        >
          {t('auth.signUp')}
        </Link>
      </p>
    </div>
  )
}

function humanizeSignInError(
  error: unknown,
  t: ReturnType<typeof useTranslation>['t'],
): string {
  const err = ConnectError.from(error)
  if (err.code === Code.Unauthenticated) {
    return t('auth.invalidCredentials')
  }
  return err.rawMessage || t('common.unexpectedError')
}
