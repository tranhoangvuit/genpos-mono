import type { TFunction } from 'i18next'
import { z } from 'zod'

export const BUSINESS_TYPES = ['fnb', 'retail', 'service', 'grocery', 'other'] as const
export type BusinessType = (typeof BUSINESS_TYPES)[number]

export function slugifyDomain(name: string): string {
  return name
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 32)
}

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
    businessName: z
      .string()
      .min(1, t('auth.validation.businessNameRequired'))
      .max(64, t('auth.validation.businessNameMax'))
      .refine((v) => slugifyDomain(v).length > 0, {
        message: t('auth.validation.businessNameSluggable'),
      }),
    email: z
      .string()
      .min(1, t('auth.validation.emailRequired'))
      .email(t('auth.validation.emailInvalid')),
    password: z
      .string()
      .min(1, t('auth.validation.passwordRequired'))
      .min(8, t('auth.validation.passwordMin')),
    businessType: z.enum(BUSINESS_TYPES),
    agreeTerms: z.boolean().refine((v) => v === true, {
      message: t('auth.validation.termsRequired'),
    }),
  })
}

export type SignUpValues = z.infer<ReturnType<typeof signUpSchema>>
