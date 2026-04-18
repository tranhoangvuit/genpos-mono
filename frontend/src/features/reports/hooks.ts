import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { useQuery as usePowerSyncQuery } from '@powersync/react'
import { useQuery } from '@tanstack/react-query'

import { orderClient } from '@/shared/api/client'
import { useAuthStore } from '@/shared/auth/store'

type DailySalesArgs = {
  dateFrom: Date
  dateTo: Date
  storeId: string // empty = all
}

const DAILY_SALES_KEY = 'reports-daily-sales'

export function useDailySalesOrders(args: DailySalesArgs) {
  return useQuery({
    queryKey: [DAILY_SALES_KEY, args.storeId, args.dateFrom.toISOString(), args.dateTo.toISOString()] as const,
    queryFn: async () => {
      const res = await orderClient.listOrders({
        storeId: args.storeId,
        dateFrom: timestampFromDate(args.dateFrom),
        dateTo: timestampFromDate(args.dateTo),
      })
      return res.orders
    },
  })
}

// useOrgTimezone reads the current org's timezone from PowerSync (synced locally).
// Returns UTC if the org row isn't loaded yet.
export function useOrgTimezone(): string {
  const orgId = useAuthStore((s) => s.user?.orgId)
  const { data } = usePowerSyncQuery<{ timezone: string }>(
    'SELECT timezone FROM organizations WHERE id = ? LIMIT 1',
    [orgId ?? ''],
  )
  return data?.[0]?.timezone || 'UTC'
}

export function useOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['reports-order', id] as const,
    enabled: !!id,
    queryFn: async () => {
      const res = await orderClient.getOrder({ id: id! })
      return res.order
    },
  })
}

// useOrgStores returns all stores for the org from PowerSync (synced locally).
export function useOrgStores() {
  const { data } = usePowerSyncQuery<{ id: string; name: string }>(
    'SELECT id, name FROM stores ORDER BY name ASC',
  )
  return data ?? []
}
