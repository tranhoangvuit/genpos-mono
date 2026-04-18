import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { taxRateClient } from '@/shared/api/client'

const KEY = ['settings', 'taxes'] as const

export function useTaxRates() {
  return useQuery({
    queryKey: KEY,
    queryFn: async () => {
      const res = await taxRateClient.listTaxRates({})
      return res.rates
    },
  })
}

export function useCreateTaxRate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof taxRateClient.createTaxRate>[0]) =>
      taxRateClient.createTaxRate(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}

export function useUpdateTaxRate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof taxRateClient.updateTaxRate>[0]) =>
      taxRateClient.updateTaxRate(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}

export function useDeleteTaxRate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => taxRateClient.deleteTaxRate({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}
