import { authClient, setOnAuthFailure } from '@/shared/api/client'
import { connectSync, disconnectSync } from '@/shared/sync/client'

import { useAuthStore } from './store'

let hydrationPromise: Promise<void> | null = null

export function bootstrapAuth(): Promise<void> {
  if (typeof window === 'undefined') {
    return Promise.resolve()
  }
  if (!hydrationPromise) {
    hydrationPromise = (async () => {
      try {
        const res = await authClient.me({})
        useAuthStore.getState().setUser(res.user ?? null)
        if (res.user) {
          void connectSync()
        }
      } catch {
        useAuthStore.getState().setUser(null)
      }
    })()
  }
  return hydrationPromise
}

export function resetAuthBootstrap(): void {
  hydrationPromise = null
  useAuthStore.getState().clear()
  void disconnectSync()
}

setOnAuthFailure(() => {
  resetAuthBootstrap()
})
