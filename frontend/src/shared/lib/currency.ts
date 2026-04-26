const ZERO_DECIMAL = new Set(['VND', 'JPY', 'KRW', 'IDR', 'CLP'])

const SYMBOLS: Record<string, string> = {
  USD: '$',
  SGD: 'S$',
  EUR: '€',
  GBP: '£',
  JPY: '¥',
  AUD: 'A$',
  VND: '₫',
}

export function currencyFractionDigits(currency: string): number {
  return ZERO_DECIMAL.has(currency.toUpperCase()) ? 0 : 2
}

export function currencySymbol(currency: string): string {
  return SYMBOLS[currency.toUpperCase()] ?? currency.toUpperCase()
}

export function formatMoney(amount: number, currency: string = 'VND'): string {
  const digits = currencyFractionDigits(currency)
  const num = Number.isFinite(amount) ? amount : 0
  const formatted = num.toLocaleString(undefined, {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })
  const sym = currencySymbol(currency)
  return currency.toUpperCase() === 'VND' ? `${formatted}${sym}` : `${sym}${formatted}`
}
