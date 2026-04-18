import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { customerClient } from '@/shared/api/client'

const CUSTOMERS_KEY = ['customers', 'list'] as const
const CUSTOMER_GROUPS_KEY = ['customers', 'groups'] as const

export function useCustomers() {
  return useQuery({
    queryKey: CUSTOMERS_KEY,
    queryFn: async () => {
      const res = await customerClient.listCustomers({})
      return res.customers
    },
  })
}

export function useCustomerGroups() {
  return useQuery({
    queryKey: CUSTOMER_GROUPS_KEY,
    queryFn: async () => {
      const res = await customerClient.listCustomerGroups({})
      return res.groups
    },
  })
}

// ----- ConnectRPC mutations ------------------------------------------------

export function useGetCustomer() {
  return useMutation({
    mutationFn: (id: string) => customerClient.getCustomer({ id }),
  })
}

export function useCreateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof customerClient.createCustomer>[0]) =>
      customerClient.createCustomer(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMERS_KEY })
    },
  })
}

export function useUpdateCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof customerClient.updateCustomer>[0]) =>
      customerClient.updateCustomer(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMERS_KEY })
    },
  })
}

export function useDeleteCustomer() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => customerClient.deleteCustomer({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMERS_KEY })
    },
  })
}

export function useCreateCustomerGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof customerClient.createCustomerGroup>[0]) =>
      customerClient.createCustomerGroup(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMER_GROUPS_KEY })
    },
  })
}

export function useUpdateCustomerGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof customerClient.updateCustomerGroup>[0]) =>
      customerClient.updateCustomerGroup(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMER_GROUPS_KEY })
    },
  })
}

export function useDeleteCustomerGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => customerClient.deleteCustomerGroup({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMER_GROUPS_KEY })
    },
  })
}

export function useParseImportCustomerCsv() {
  return useMutation({
    mutationFn: (csvData: Uint8Array) =>
      customerClient.parseImportCustomerCsv({ csvData }),
  })
}

export function useImportCustomers() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof customerClient.importCustomers>[0]) =>
      customerClient.importCustomers(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: CUSTOMERS_KEY })
    },
  })
}
