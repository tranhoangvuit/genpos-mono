import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { catalogClient } from '@/shared/api/client'
import type { CreateCategoryRequest, UpdateCategoryRequest } from '@/gen/genpos/v1/catalog_pb'

const PRODUCTS_KEY = ['catalog', 'products'] as const
const CATEGORIES_KEY = ['catalog', 'categories'] as const

export function useCategories() {
  return useQuery({
    queryKey: CATEGORIES_KEY,
    queryFn: async () => {
      const res = await catalogClient.listCategories({})
      return res.categories
    },
  })
}

export function useProductList() {
  return useQuery({
    queryKey: PRODUCTS_KEY,
    queryFn: async () => {
      const res = await catalogClient.listProducts({})
      return res.products
    },
  })
}

// ----- Mutations -----------------------------------------------------------

export function useCreateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<CreateCategoryRequest>) =>
      catalogClient.createCategory({
        name: input.name ?? '',
        parentId: input.parentId ?? '',
        color: input.color ?? '',
        sortOrder: input.sortOrder ?? 0,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CATEGORIES_KEY })
    },
  })
}

export function useUpdateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<UpdateCategoryRequest> & { id: string }) =>
      catalogClient.updateCategory({
        id: input.id,
        name: input.name ?? '',
        parentId: input.parentId ?? '',
        color: input.color ?? '',
        sortOrder: input.sortOrder ?? 0,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CATEGORIES_KEY })
    },
  })
}

export function useDeleteCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => catalogClient.deleteCategory({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CATEGORIES_KEY })
    },
  })
}

export function useCreateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof catalogClient.createProduct>[0]) =>
      catalogClient.createProduct(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PRODUCTS_KEY })
    },
  })
}

export function useUpdateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof catalogClient.updateProduct>[0]) =>
      catalogClient.updateProduct(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PRODUCTS_KEY })
    },
  })
}

export function useDeleteProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => catalogClient.deleteProduct({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PRODUCTS_KEY })
    },
  })
}

export function useGetProduct() {
  return useMutation({
    mutationFn: (id: string) => catalogClient.getProduct({ id }),
  })
}

export function useParseImportCsv() {
  return useMutation({
    mutationFn: (csvData: Uint8Array) => catalogClient.parseImportCsv({ csvData }),
  })
}

export function useImportProducts() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof catalogClient.importProducts>[0]) =>
      catalogClient.importProducts(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: PRODUCTS_KEY })
    },
  })
}
