import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import vn from './locales/vn.json'

export const SUPPORTED_LANGUAGES = ['en', 'vn'] as const
export type Language = (typeof SUPPORTED_LANGUAGES)[number]

const STORAGE_KEY = 'app-language'

function initialLanguage(): Language {
  if (typeof window === 'undefined') return 'en'
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored && (SUPPORTED_LANGUAGES as readonly string[]).includes(stored)) {
    return stored as Language
  }
  return 'en'
}

if (!i18n.isInitialized) {
  void i18n.use(initReactI18next).init({
    resources: {
      en: { translation: en },
      vn: { translation: vn },
    },
    lng: initialLanguage(),
    fallbackLng: 'en',
    interpolation: { escapeValue: false },
    returnNull: false,
  })
}

export function setLanguage(lang: Language): void {
  void i18n.changeLanguage(lang)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, lang)
  }
}

export function getLanguage(): Language {
  return (i18n.language as Language) || 'en'
}

export default i18n
