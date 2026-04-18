import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  purchaseOrderClient,
  stockTakeClient,
  supplierClient,
} from '@/shared/api/client'

const SUPPLIERS_KEY = ['inventory', 'suppliers'] as const
const PURCHASE_ORDERS_KEY = ['inventory', 'purchase-orders'] as const
const STORES_KEY = ['inventory', 'stores'] as const
const VARIANT_PICKER_KEY = ['inventory', 'variant-picker'] as const
const STOCK_TAKES_KEY = ['inventory', 'stock-takes'] as const

// ----- Reads ---------------------------------------------------------------

export function useSuppliers() {
  return useQuery({
    queryKey: SUPPLIERS_KEY,
    queryFn: async () => {
      const res = await supplierClient.listSuppliers({})
      return res.suppliers
    },
  })
}

export function usePurchaseOrders() {
  return useQuery({
    queryKey: PURCHASE_ORDERS_KEY,
    queryFn: async () => {
      const res = await purchaseOrderClient.listPurchaseOrders({})
      return res.purchaseOrders
    },
  })
}

export function useStores() {
  return useQuery({
    queryKey: STORES_KEY,
    queryFn: async () => {
      const res = await purchaseOrderClient.listStores({})
      return res.stores
    },
  })
}

export function useVariantPicker() {
  return useQuery({
    queryKey: VARIANT_PICKER_KEY,
    queryFn: async () => {
      const res = await purchaseOrderClient.listVariantsForPicker({})
      return res.variants
    },
  })
}

export function useStockTakes() {
  return useQuery({
    queryKey: STOCK_TAKES_KEY,
    queryFn: async () => {
      const res = await stockTakeClient.listStockTakes({})
      return res.stockTakes
    },
  })
}

// ----- ConnectRPC mutations ------------------------------------------------

export function useCreateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof supplierClient.createSupplier>[0]) =>
      supplierClient.createSupplier(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: SUPPLIERS_KEY })
    },
  })
}

export function useUpdateSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof supplierClient.updateSupplier>[0]) =>
      supplierClient.updateSupplier(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: SUPPLIERS_KEY })
    },
  })
}

export function useDeleteSupplier() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => supplierClient.deleteSupplier({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: SUPPLIERS_KEY })
    },
  })
}

export function useGetPurchaseOrder() {
  return useMutation({
    mutationFn: (id: string) => purchaseOrderClient.getPurchaseOrder({ id }),
  })
}

export function usePurchaseOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['inventory', 'purchase-order', id] as const,
    enabled: !!id,
    queryFn: async () => {
      const res = await purchaseOrderClient.getPurchaseOrder({ id: id! })
      return res.purchaseOrder
    },
  })
}

export function useStockTake(id: string | undefined) {
  return useQuery({
    queryKey: ['inventory', 'stock-take', id] as const,
    enabled: !!id,
    queryFn: async () => {
      const res = await stockTakeClient.getStockTake({ id: id! })
      return res.stockTake
    },
  })
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof purchaseOrderClient.createPurchaseOrder>[0]) =>
      purchaseOrderClient.createPurchaseOrder(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

export function useUpdatePurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof purchaseOrderClient.updatePurchaseOrder>[0]) =>
      purchaseOrderClient.updatePurchaseOrder(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

export function useSubmitPurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => purchaseOrderClient.submitPurchaseOrder({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

export function useReceivePurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof purchaseOrderClient.receivePurchaseOrder>[0]) =>
      purchaseOrderClient.receivePurchaseOrder(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

export function useCancelPurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => purchaseOrderClient.cancelPurchaseOrder({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

export function useDeletePurchaseOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => purchaseOrderClient.deletePurchaseOrder({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PURCHASE_ORDERS_KEY })
    },
  })
}

// ----- Stock takes ---------------------------------------------------------

export function useGetStockTake() {
  return useMutation({
    mutationFn: (id: string) => stockTakeClient.getStockTake({ id }),
  })
}

export function useCreateStockTake() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof stockTakeClient.createStockTake>[0]) =>
      stockTakeClient.createStockTake(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STOCK_TAKES_KEY })
    },
  })
}

export function useSaveStockTakeProgress() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof stockTakeClient.saveStockTakeProgress>[0]) =>
      stockTakeClient.saveStockTakeProgress(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STOCK_TAKES_KEY })
    },
  })
}

export function useFinalizeStockTake() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stockTakeClient.finalizeStockTake({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STOCK_TAKES_KEY })
    },
  })
}

export function useCancelStockTake() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stockTakeClient.cancelStockTake({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STOCK_TAKES_KEY })
    },
  })
}

export function useDeleteStockTake() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stockTakeClient.deleteStockTake({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STOCK_TAKES_KEY })
    },
  })
}
