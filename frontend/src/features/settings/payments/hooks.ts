import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { paymentMethodClient } from '@/shared/api/client'

const KEY = ['settings', 'payments'] as const

export function usePaymentMethods() {
  return useQuery({
    queryKey: KEY,
    queryFn: async () => {
      const res = await paymentMethodClient.listPaymentMethods({})
      return res.methods
    },
  })
}

export function useCreatePaymentMethod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof paymentMethodClient.createPaymentMethod>[0]) =>
      paymentMethodClient.createPaymentMethod(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}

export function useUpdatePaymentMethod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof paymentMethodClient.updatePaymentMethod>[0]) =>
      paymentMethodClient.updatePaymentMethod(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}

export function useDeletePaymentMethod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => paymentMethodClient.deletePaymentMethod({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: KEY })
    },
  })
}
