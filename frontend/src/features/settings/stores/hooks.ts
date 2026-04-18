import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { storeClient } from '@/shared/api/client'

const STORES_KEY = ['settings', 'stores'] as const

export function useStores() {
  return useQuery({
    queryKey: STORES_KEY,
    queryFn: async () => {
      const res = await storeClient.listStoreDetails({})
      return res.stores
    },
  })
}

export function useCreateStore() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof storeClient.createStore>[0]) =>
      storeClient.createStore(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STORES_KEY })
    },
  })
}

export function useUpdateStore() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof storeClient.updateStore>[0]) =>
      storeClient.updateStore(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STORES_KEY })
    },
  })
}

export function useDeleteStore() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => storeClient.deleteStore({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STORES_KEY })
    },
  })
}
