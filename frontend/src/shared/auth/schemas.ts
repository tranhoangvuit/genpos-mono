import type { TFunction } from 'i18next'
import { z } from 'zod'

const DOMAIN_REGEX = /^[a-z0-9-]{1,32}$/

export function signInSchema(t: TFunction) {
  return z.object({
    email: z
      .string()
      .min(1, t('auth.validation.emailRequired'))
      .email(t('auth.validation.emailInvalid')),
    password: z
      .string()
      .min(1, t('auth.validation.passwordRequired'))
      .min(8, t('auth.validation.passwordMin')),
    rememberMe: z.boolean(),
  })
}

export type SignInValues = z.infer<ReturnType<typeof signInSchema>>

export function signUpSchema(t: TFunction) {
  return z.object({
    domain: z
      .string()
      .min(1, t('auth.validation.domainRequired'))
      .regex(DOMAIN_REGEX, t('auth.validation.domainInvalid')),
    email: z
      .string()
      .min(1, t('auth.validation.emailRequired'))
      .email(t('auth.validation.emailInvalid')),
    password: z
      .string()
      .min(1, t('auth.validation.passwordRequired'))
      .min(8, t('auth.validation.passwordMin')),
  })
}

export type SignUpValues = z.infer<ReturnType<typeof signUpSchema>>
