import type { OptionFormValue, VariantFormValue } from './types'

export const emptyVariant = (): VariantFormValue => ({
  name: 'Default',
  sku: '',
  barcode: '',
  price: '0',
  costPrice: '0',
  trackStock: true,
  isActive: true,
  sortOrder: 0,
  optionValues: [],
})

// cartesian returns the cartesian product of option value arrays.
// Each row is a tuple ordered to match the options array.
function cartesian(arrays: string[][]): string[][] {
  if (arrays.length === 0) return []
  return arrays.reduce<string[][]>(
    (acc, cur) => acc.flatMap((a) => cur.map((b) => [...a, b])),
    [[]],
  )
}

// generateVariants builds the full variant list from options, preserving
// previously-edited variants when their option combination still exists.
export function generateVariants(
  options: OptionFormValue[],
  existing: VariantFormValue[],
): VariantFormValue[] {
  const cleaned = options.filter((o) => o.name.trim() && o.values.length > 0)
  if (cleaned.length === 0) return existing.length > 0 ? existing : [emptyVariant()]

  const valueArrays = cleaned.map((o) => o.values)
  const tuples = cartesian(valueArrays)

  const keyOf = (values: string[]) => JSON.stringify(values)
  const existingByKey = new Map<string, VariantFormValue>()
  for (const v of existing) {
    existingByKey.set(keyOf(v.optionValues), v)
  }

  return tuples.map((tuple, i) => {
    const key = keyOf(tuple)
    const prior = existingByKey.get(key)
    return {
      ...emptyVariant(),
      ...(prior ?? {}),
      name: tuple.join(' / '),
      optionValues: tuple,
      sortOrder: i,
    }
  })
}
