import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { ConnectError } from '@connectrpc/connect'
import {
  ArrowRight,
  ArrowUpCircle,
  Building2,
  Check,
  CircleHelp,
  Eye,
  EyeOff,
  Lock,
  Mail,
  ShieldCheck,
  ShoppingBag,
  ShoppingCart,
  UtensilsCrossed,
  Wrench,
} from 'lucide-react'
import { useMemo, useState, type ComponentType, type SVGProps } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { useSignUp } from '@/shared/auth/hooks'
import {
  type BusinessType,
  signUpSchema,
  slugifyDomain,
  type SignUpValues,
} from '@/shared/auth/schemas'

type ChipDef = {
  value: BusinessType
  label: string
  hint: string
  icon: ComponentType<SVGProps<SVGSVGElement>>
}

const PRIMARY = 'hsl(221 83% 53%)'
const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(215.4 16.3% 46.9%)'
const FOREGROUND = 'hsl(222.2 84% 4.9%)'
const DESTRUCTIVE = 'hsl(0 84% 60%)'

export function RegisterCard() {
  const { t } = useTranslation()
  const signUp = useSignUp()
  const [showPassword, setShowPassword] = useState(false)
  const [capsOn, setCapsOn] = useState(false)

  const form = useForm<SignUpValues>({
    resolver: standardSchemaResolver(signUpSchema(t)),
    defaultValues: {
      businessName: '',
      email: '',
      password: '',
      businessType: 'fnb',
      agreeTerms: true,
    },
  })

  const password = form.watch('password')
  const businessName = form.watch('businessName')
  const slug = useMemo(() => slugifyDomain(businessName ?? ''), [businessName])

  const onSubmit = form.handleSubmit((values) => {
    signUp.mutate({
      domain: slugifyDomain(values.businessName),
      email: values.email,
      password: values.password,
      businessType: values.businessType,
    })
  })

  const serverMessage = signUp.error
    ? ConnectError.from(signUp.error).rawMessage || t('common.unexpectedError')
    : null

  const strength = passwordStrength(password ?? '', t)
  const chips: ChipDef[] = [
    { value: 'fnb', label: t('auth.bizType.fnb'), hint: t('auth.bizType.fnbHint'), icon: UtensilsCrossed },
    { value: 'retail', label: t('auth.bizType.retail'), hint: t('auth.bizType.retailHint'), icon: ShoppingBag },
    { value: 'service', label: t('auth.bizType.service'), hint: t('auth.bizType.serviceHint'), icon: Wrench },
    { value: 'grocery', label: t('auth.bizType.grocery'), hint: t('auth.bizType.groceryHint'), icon: ShoppingCart },
    { value: 'other', label: t('auth.bizType.other'), hint: t('auth.bizType.otherHint'), icon: CircleHelp },
  ]

  return (
    <div className="w-full">
      <h2 className="m-0 mb-1.5 text-[30px] font-bold tracking-[-0.02em]" style={{ color: FOREGROUND }}>
        {t('auth.signUpTitle')}
      </h2>
      <p className="m-0 mb-7 text-[14px]" style={{ color: MUTED }}>
        {t('auth.signUpSubtitle')}
      </p>

      {serverMessage && (
        <div
          role="alert"
          className="mb-4 rounded-md border px-3 py-2 text-sm"
          style={{
            borderColor: 'hsl(0 84% 60% / 0.3)',
            background: 'hsl(0 84% 60% / 0.1)',
            color: 'hsl(0 84% 40%)',
          }}
        >
          {serverMessage}
        </div>
      )}

      <form onSubmit={onSubmit} noValidate>
        {/* Business name */}
        <Field
          id="businessName"
          label={t('auth.businessName')}
          error={form.formState.errors.businessName?.message}
          hint={
            slug
              ? t('auth.businessDomainPreview', { slug })
              : t('auth.businessNameHint')
          }
        >
          <InputWithIcon
            id="businessName"
            type="text"
            placeholder={t('auth.businessNamePlaceholder')}
            autoComplete="organization"
            aria-invalid={form.formState.errors.businessName ? true : undefined}
            icon={Building2}
            {...form.register('businessName')}
          />
        </Field>

        {/* Email */}
        <Field
          id="email"
          label={t('auth.workEmail')}
          error={form.formState.errors.email?.message}
        >
          <InputWithIcon
            id="email"
            type="email"
            placeholder="you@business.com"
            autoComplete="email"
            aria-invalid={form.formState.errors.email ? true : undefined}
            icon={Mail}
            {...form.register('email')}
          />
        </Field>

        {/* Password */}
        <div className="mb-3.5">
          <label
            htmlFor="password"
            className="mb-1.5 block text-[12.5px] font-medium"
            style={{ color: FOREGROUND }}
          >
            {t('auth.password')}
          </label>
          <div className="relative">
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2" style={{ color: MUTED }}>
              <Lock className="h-[15px] w-[15px]" strokeWidth={2} />
            </span>
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              autoComplete="new-password"
              placeholder={t('auth.passwordPlaceholder')}
              aria-invalid={form.formState.errors.password ? true : undefined}
              {...form.register('password')}
              onKeyDown={(e) => setCapsOn(!!e.getModifierState?.('CapsLock'))}
              onKeyUp={(e) => setCapsOn(!!e.getModifierState?.('CapsLock'))}
              className="h-11 w-full rounded-md border bg-white pl-[38px] pr-10 text-[14px] outline-none transition focus:shadow-[0_0_0_3px_hsl(221_83%_53%_/_0.15)]"
              style={{
                color: FOREGROUND,
                borderColor: form.formState.errors.password ? DESTRUCTIVE : BORDER,
              }}
              onFocus={(e) => {
                if (!form.formState.errors.password) {
                  e.currentTarget.style.borderColor = PRIMARY
                }
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = form.formState.errors.password
                  ? DESTRUCTIVE
                  : BORDER
              }}
            />
            <button
              type="button"
              onClick={() => setShowPassword((s) => !s)}
              tabIndex={-1}
              aria-label={showPassword ? t('auth.hidePassword') : t('auth.showPassword')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-[hsl(210_40%_96%)]"
              style={{ color: MUTED }}
            >
              {showPassword ? (
                <EyeOff className="h-[15px] w-[15px]" strokeWidth={2} />
              ) : (
                <Eye className="h-[15px] w-[15px]" strokeWidth={2} />
              )}
            </button>
          </div>

          {/* Password strength meter */}
          <div className="mt-2 flex items-center gap-2.5">
            <div
              className="relative h-1 flex-1 overflow-hidden rounded-full"
              style={{ background: 'hsl(210 40% 96.1%)' }}
            >
              <div
                className="h-full rounded-full transition-all duration-200"
                style={{
                  width: strength.width,
                  background: strength.color,
                }}
              />
            </div>
            <span
              className="whitespace-nowrap text-[11px] tabular-nums"
              style={{ color: password ? strength.color : MUTED }}
            >
              {password ? strength.label : t('auth.passwordHint')}
            </span>
          </div>

          {capsOn && (
            <div
              className="mt-1.5 inline-flex items-center gap-1.5 rounded px-2.5 py-1.5 text-[11.5px]"
              style={{
                background: 'hsl(33.3 100% 96.5%)',
                color: 'hsl(22.7 82.5% 45.1%)',
              }}
            >
              <ArrowUpCircle className="h-3 w-3" strokeWidth={2} />
              {t('auth.capsLockOn')}
            </div>
          )}
          {form.formState.errors.password && (
            <p className="mt-1 text-[12px]" style={{ color: DESTRUCTIVE }}>
              {form.formState.errors.password.message}
            </p>
          )}
        </div>

        {/* Business type chips */}
        <div className="mb-3.5">
          <label className="mb-1.5 block text-[12.5px] font-medium" style={{ color: FOREGROUND }}>
            {t('auth.businessType')}
          </label>
          <Controller
            control={form.control}
            name="businessType"
            render={({ field }) => (
              <div role="radiogroup" aria-label={t('auth.businessType')} className="grid grid-cols-3 gap-2">
                {chips.map((chip) => {
                  const active = field.value === chip.value
                  const Icon = chip.icon
                  return (
                    <button
                      key={chip.value}
                      type="button"
                      role="radio"
                      aria-checked={active}
                      onClick={() => field.onChange(chip.value)}
                      className="relative flex min-h-[72px] flex-col items-start gap-1 rounded-md border px-2.5 py-2.5 text-left transition active:translate-y-px"
                      style={{
                        background: active ? 'hsl(221 83% 53% / 0.06)' : 'white',
                        borderColor: active ? PRIMARY : BORDER,
                        boxShadow: active ? '0 0 0 3px hsl(221 83% 53% / 0.12)' : undefined,
                      }}
                    >
                      <Icon
                        className="h-4 w-4"
                        strokeWidth={1.8}
                        style={{ color: active ? PRIMARY : MUTED }}
                      />
                      <span
                        className="text-[12.5px] font-semibold tracking-[-0.005em]"
                        style={{ color: active ? PRIMARY : FOREGROUND }}
                      >
                        {chip.label}
                      </span>
                      <span
                        className="text-[10.5px] font-medium leading-tight"
                        style={{ color: active ? 'hsl(221 83% 53% / 0.8)' : MUTED }}
                      >
                        {chip.hint}
                      </span>
                      {active && (
                        <span
                          aria-hidden
                          className="absolute right-2 top-2 grid h-[14px] w-[14px] place-items-center rounded-full"
                          style={{ background: PRIMARY }}
                        >
                          <Check className="h-2.5 w-2.5 text-white" strokeWidth={4} />
                        </span>
                      )}
                    </button>
                  )
                })}
              </div>
            )}
          />
        </div>

        {/* Terms */}
        <div className="my-[18px]">
          <Controller
            control={form.control}
            name="agreeTerms"
            render={({ field }) => {
              const checked = !!field.value
              return (
                <label className="inline-flex cursor-pointer select-none items-center gap-2 text-[13px]" style={{ color: FOREGROUND }}>
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
                      borderColor: checked ? PRIMARY : BORDER,
                      background: checked ? PRIMARY : 'white',
                    }}
                  >
                    {checked && (
                      <Check className="h-[11px] w-[11px] text-white" strokeWidth={3.5} />
                    )}
                  </span>
                  <span>
                    {t('auth.agreeTo')}{' '}
                    <a href="#" onClick={(e) => e.preventDefault()} className="font-medium hover:underline" style={{ color: PRIMARY }}>
                      {t('auth.terms')}
                    </a>{' '}
                    &amp;{' '}
                    <a href="#" onClick={(e) => e.preventDefault()} className="font-medium hover:underline" style={{ color: PRIMARY }}>
                      {t('auth.privacy')}
                    </a>
                  </span>
                </label>
              )
            }}
          />
          {form.formState.errors.agreeTerms && (
            <p className="mt-1 text-[12px]" style={{ color: DESTRUCTIVE }}>
              {form.formState.errors.agreeTerms.message}
            </p>
          )}
        </div>

        <button
          type="submit"
          disabled={signUp.isPending}
          className="inline-flex h-11 w-full items-center justify-center gap-2 rounded-md text-[14px] font-semibold text-white transition disabled:cursor-not-allowed disabled:opacity-70"
          style={{
            background: PRIMARY,
            boxShadow: '0 4px 12px hsl(221 83% 53% / 0.25)',
          }}
          onMouseEnter={(e) => {
            if (!signUp.isPending) e.currentTarget.style.background = 'hsl(221 83% 48%)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = PRIMARY
          }}
        >
          {signUp.isPending ? t('auth.creatingAccount') : t('auth.createAccount')}
          {!signUp.isPending && <ArrowRight className="h-[15px] w-[15px]" strokeWidth={2.5} />}
        </button>

        <div className="mt-6 flex items-center gap-1.5 border-t pt-4 text-[12.5px]" style={{ borderColor: BORDER, color: MUTED }}>
          <ShieldCheck className="h-[13px] w-[13px]" strokeWidth={2} style={{ color: 'hsl(142.1 76.2% 36.3%)' }} />
          {t('auth.trialFooter')}
        </div>
      </form>
    </div>
  )
}

function Field({
  id,
  label,
  hint,
  error,
  children,
}: {
  id: string
  label: string
  hint?: string
  error?: string
  children: React.ReactNode
}) {
  return (
    <div className="mb-3.5">
      <label htmlFor={id} className="mb-1.5 block text-[12.5px] font-medium" style={{ color: FOREGROUND }}>
        {label}
      </label>
      {children}
      {error ? (
        <p className="mt-1 text-[12px]" style={{ color: DESTRUCTIVE }}>
          {error}
        </p>
      ) : hint ? (
        <p className="mt-1 text-[11.5px]" style={{ color: MUTED }}>
          {hint}
        </p>
      ) : null}
    </div>
  )
}

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  icon: ComponentType<SVGProps<SVGSVGElement>>
}

const InputWithIcon = function InputWithIcon({ icon: Icon, style, ...props }: InputProps) {
  const [focused, setFocused] = useState(false)
  const isInvalid = props['aria-invalid'] === true
  return (
    <div className="relative">
      <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2" style={{ color: MUTED }}>
        <Icon className="h-[15px] w-[15px]" strokeWidth={2} />
      </span>
      <input
        {...props}
        onFocus={(e) => {
          setFocused(true)
          props.onFocus?.(e)
        }}
        onBlur={(e) => {
          setFocused(false)
          props.onBlur?.(e)
        }}
        className="h-11 w-full rounded-md border bg-white pl-[38px] pr-3 text-[14px] outline-none transition placeholder:text-[hsl(215.4_16.3%_46.9%)] focus:shadow-[0_0_0_3px_hsl(221_83%_53%_/_0.15)]"
        style={{
          color: FOREGROUND,
          borderColor: isInvalid ? DESTRUCTIVE : focused ? PRIMARY : BORDER,
          ...style,
        }}
      />
    </div>
  )
}

type StrengthLevel = 'tooShort' | 'weak' | 'fair' | 'good' | 'strong'
type Strength = { width: string; color: string; label: string }

const STRENGTH_TONES: Array<{ width: string; color: string; level: StrengthLevel }> = [
  { width: '8%', color: 'hsl(24.6 95% 53.1%)', level: 'tooShort' },
  { width: '28%', color: 'hsl(24.6 95% 53.1%)', level: 'weak' },
  { width: '55%', color: 'hsl(35 95% 55%)', level: 'fair' },
  { width: '80%', color: PRIMARY, level: 'good' },
  { width: '100%', color: 'hsl(142.1 76.2% 36.3%)', level: 'strong' },
]

function passwordStrength(
  v: string,
  t: (key: string) => string,
): Strength {
  if (!v) return { width: '0%', color: 'hsl(215 15% 60%)', label: '' }
  let score = 0
  if (v.length >= 8) score++
  if (v.length >= 12) score++
  if (/[A-Z]/.test(v) && /[a-z]/.test(v)) score++
  if (/\d/.test(v)) score++
  if (/[^A-Za-z0-9]/.test(v)) score++
  score = Math.min(score, 4)
  const tone = STRENGTH_TONES[score]
  return {
    width: tone.width,
    color: tone.color,
    label: t(`auth.passwordStrength_${tone.level}`),
  }
}
