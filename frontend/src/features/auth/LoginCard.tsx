import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { Code, ConnectError } from '@connectrpc/connect'
import { Link } from '@tanstack/react-router'
import { ArrowRight, ArrowUpCircle, Eye, EyeOff, Lock, Mail, ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { useSignIn } from '@/shared/auth/hooks'
import { signInSchema, type SignInValues } from '@/shared/auth/schemas'

export function LoginCard() {
  const { t } = useTranslation()
  const signIn = useSignIn()
  const [showPassword, setShowPassword] = useState(false)
  const [capsOn, setCapsOn] = useState(false)

  const form = useForm<SignInValues>({
    resolver: standardSchemaResolver(signInSchema(t)),
    defaultValues: { email: '', password: '', rememberMe: true },
  })

  const onSubmit = form.handleSubmit((values) => {
    signIn.mutate(values)
  })

  const serverMessage = signIn.error
    ? humanizeSignInError(signIn.error, t)
    : null

  return (
    <div className="w-full">
      <h2 className="m-0 mb-1.5 text-[30px] font-bold tracking-[-0.02em] text-[hsl(222.2_84%_4.9%)]">
        Welcome back
      </h2>
      <p className="m-0 mb-7 text-[14px] text-[hsl(215.4_16.3%_46.9%)]">
        Sign in to your register and pick up where you left off.
      </p>

      {serverMessage && (
        <div
          role="alert"
          className="mb-4 rounded-md border border-[hsl(0_84%_60%_/_0.3)] bg-[hsl(0_84%_60%_/_0.1)] px-3 py-2 text-sm text-[hsl(0_84%_40%)]"
        >
          {serverMessage}
        </div>
      )}

      <form onSubmit={onSubmit} noValidate>
        <div className="mb-3.5">
          <label
            htmlFor="email"
            className="mb-1.5 block text-[12.5px] font-medium text-[hsl(222.2_84%_4.9%)]"
          >
            Work email
          </label>
          <div className="relative">
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[hsl(215.4_16.3%_46.9%)]">
              <Mail className="h-[15px] w-[15px]" strokeWidth={2} />
            </span>
            <input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="maya@bluebird.coffee"
              aria-invalid={form.formState.errors.email ? true : undefined}
              {...form.register('email')}
              className={`h-11 w-full rounded-md border bg-white pl-[38px] pr-3 text-[14px] text-[hsl(222.2_84%_4.9%)] outline-none transition placeholder:text-[hsl(215.4_16.3%_46.9%)] focus:border-[hsl(221_83%_53%)] focus:shadow-[0_0_0_3px_hsl(221_83%_53%_/_0.15)] ${
                form.formState.errors.email
                  ? 'border-[hsl(0_84%_60%)]'
                  : 'border-[hsl(214.3_31.8%_91.4%)]'
              }`}
            />
          </div>
          {form.formState.errors.email && (
            <p className="mt-1 text-[12px] text-[hsl(0_84%_60%)]">
              {form.formState.errors.email.message}
            </p>
          )}
        </div>

        <div className="mb-3.5">
          <label
            htmlFor="password"
            className="mb-1.5 flex items-center justify-between text-[12.5px] font-medium text-[hsl(222.2_84%_4.9%)]"
          >
            {t('auth.password')}
            <a
              href="#"
              onClick={(e) => e.preventDefault()}
              className="text-[12px] font-medium text-[hsl(221_83%_53%)] hover:underline"
            >
              Forgot?
            </a>
          </label>
          <div className="relative">
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[hsl(215.4_16.3%_46.9%)]">
              <Lock className="h-[15px] w-[15px]" strokeWidth={2} />
            </span>
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              autoComplete="current-password"
              placeholder="Enter your password"
              aria-invalid={form.formState.errors.password ? true : undefined}
              {...form.register('password')}
              onKeyDown={(e) =>
                setCapsOn(!!e.getModifierState?.('CapsLock'))
              }
              onKeyUp={(e) => setCapsOn(!!e.getModifierState?.('CapsLock'))}
              className={`h-11 w-full rounded-md border bg-white pl-[38px] pr-10 text-[14px] text-[hsl(222.2_84%_4.9%)] outline-none transition placeholder:text-[hsl(215.4_16.3%_46.9%)] focus:border-[hsl(221_83%_53%)] focus:shadow-[0_0_0_3px_hsl(221_83%_53%_/_0.15)] ${
                form.formState.errors.password
                  ? 'border-[hsl(0_84%_60%)]'
                  : 'border-[hsl(214.3_31.8%_91.4%)]'
              }`}
            />
            <button
              type="button"
              onClick={() => setShowPassword((s) => !s)}
              tabIndex={-1}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded p-1 text-[hsl(215.4_16.3%_46.9%)] hover:bg-[hsl(210_40%_96%)] hover:text-[hsl(222.2_84%_4.9%)]"
            >
              {showPassword ? (
                <EyeOff className="h-[15px] w-[15px]" strokeWidth={2} />
              ) : (
                <Eye className="h-[15px] w-[15px]" strokeWidth={2} />
              )}
            </button>
          </div>
          {capsOn && (
            <div className="mt-1.5 inline-flex items-center gap-1.5 rounded bg-[hsl(33.3_100%_96.5%)] px-2.5 py-1.5 text-[11.5px] text-[hsl(22.7_82.5%_45.1%)]">
              <ArrowUpCircle className="h-3 w-3" strokeWidth={2} />
              Caps Lock is on
            </div>
          )}
          {form.formState.errors.password && (
            <p className="mt-1 text-[12px] text-[hsl(0_84%_60%)]">
              {form.formState.errors.password.message}
            </p>
          )}
        </div>

        <div className="my-[18px] flex items-center justify-between">
          <Controller
            control={form.control}
            name="rememberMe"
            render={({ field }) => {
              const checked = !!field.value
              return (
                <label className="inline-flex cursor-pointer select-none items-center gap-2 text-[13px] text-[hsl(222.2_84%_4.9%)]">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={(e) => field.onChange(e.target.checked)}
                    className="sr-only"
                  />
                  <span
                    aria-hidden
                    className="grid h-4 w-4 place-items-center rounded border-[1.5px] transition-colors"
                    style={{
                      borderColor: checked
                        ? 'hsl(221 83% 53%)'
                        : 'hsl(214.3 31.8% 91.4%)',
                      background: checked ? 'hsl(221 83% 53%)' : 'white',
                    }}
                  >
                    {checked && (
                      <svg
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth={3.5}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="h-[11px] w-[11px] text-white"
                      >
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    )}
                  </span>
                  Keep me signed in on this register
                </label>
              )
            }}
          />
        </div>

        <button
          type="submit"
          disabled={signIn.isPending}
          className="inline-flex h-11 w-full items-center justify-center gap-2 rounded-md text-[14px] font-semibold text-white transition disabled:cursor-not-allowed disabled:opacity-70"
          style={{
            background: 'hsl(221 83% 53%)',
            boxShadow: '0 4px 12px hsl(221 83% 53% / 0.25)',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'hsl(221 83% 48%)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'hsl(221 83% 53%)'
          }}
        >
          {signIn.isPending ? t('auth.signingIn') : 'Sign in to GenPos'}
          {!signIn.isPending && (
            <ArrowRight className="h-[15px] w-[15px]" strokeWidth={2.5} />
          )}
        </button>

        <p className="mt-5 text-center text-[13px] text-[hsl(215.4_16.3%_46.9%)]">
          {t('auth.noAccount')}{' '}
          <Link
            to="/signup"
            className="font-semibold text-[hsl(221_83%_53%)] hover:underline"
          >
            {t('auth.signUp')}
          </Link>
        </p>

        <div className="mt-6 flex items-center gap-1.5 border-t border-[hsl(214.3_31.8%_91.4%)] pt-4 text-[12.5px] text-[hsl(215.4_16.3%_46.9%)]">
          <ShieldCheck
            className="h-[13px] w-[13px] text-[hsl(142.1_76.2%_36.3%)]"
            strokeWidth={2}
          />
          Protected by SOC 2 · PCI-DSS Level 1
        </div>
      </form>
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
