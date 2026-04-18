/// <reference types="vite/client" />

import {
  Code,
  ConnectError,
  createClient,
  type Interceptor,
} from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { CatalogService } from '@/gen/genpos/v1/catalog_pb'
import { CustomerService } from '@/gen/genpos/v1/customer_pb'
import { AuthService, GenposService } from '@/gen/genpos/v1/genpos_pb'
import {
  PurchaseOrderService,
  SupplierService,
} from '@/gen/genpos/v1/inventory_pb'
import { MemberService } from '@/gen/genpos/v1/member_pb'
import { OrderService } from '@/gen/genpos/v1/order_pb'
import { PaymentMethodService } from '@/gen/genpos/v1/payment_pb'
import { StockTakeService } from '@/gen/genpos/v1/stock_take_pb'
import { StoreService } from '@/gen/genpos/v1/store_pb'
import { TaxRateService } from '@/gen/genpos/v1/tax_pb'

const baseUrl =
  (import.meta.env.VITE_API_BASE_URL as string | undefined) ??
  'http://localhost:3031'

const fetchWithCredentials: typeof fetch = (input, init) =>
  fetch(input, { ...init, credentials: 'include' })

const publicAuthMethods: ReadonlySet<unknown> = new Set([
  AuthService.method.signUp,
  AuthService.method.signIn,
  AuthService.method.signOut,
  AuthService.method.refresh,
])

let onAuthFailure: (() => void) | null = null
let pendingRefresh: Promise<void> | null = null

export function setOnAuthFailure(fn: (() => void) | null): void {
  onAuthFailure = fn
}

const rawTransport = createConnectTransport({
  baseUrl,
  fetch: fetchWithCredentials,
})

const refreshBackchannel = createClient(AuthService, rawTransport)

async function performRefresh(): Promise<void> {
  if (!pendingRefresh) {
    pendingRefresh = (async () => {
      try {
        await refreshBackchannel.refresh({})
      } finally {
        pendingRefresh = null
      }
    })()
  }
  return pendingRefresh
}

const refreshInterceptor: Interceptor = (next) => async (req) => {
  try {
    return await next(req)
  } catch (err) {
    const code = ConnectError.from(err).code
    if (code !== Code.Unauthenticated || publicAuthMethods.has(req.method)) {
      throw err
    }
    try {
      await performRefresh()
    } catch {
      onAuthFailure?.()
      throw err
    }
    return await next(req)
  }
}

const transport = createConnectTransport({
  baseUrl,
  fetch: fetchWithCredentials,
  interceptors: [refreshInterceptor],
})

export const authClient = createClient(AuthService, transport)
export const genposClient = createClient(GenposService, transport)
export const catalogClient = createClient(CatalogService, transport)
export const customerClient = createClient(CustomerService, transport)
export const supplierClient = createClient(SupplierService, transport)
export const purchaseOrderClient = createClient(PurchaseOrderService, transport)
export const stockTakeClient = createClient(StockTakeService, transport)
export const storeClient = createClient(StoreService, transport)
export const paymentMethodClient = createClient(PaymentMethodService, transport)
export const taxRateClient = createClient(TaxRateService, transport)
export const memberClient = createClient(MemberService, transport)
export const orderClient = createClient(OrderService, transport)
